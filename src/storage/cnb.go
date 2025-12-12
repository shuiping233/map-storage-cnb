package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"map-storage-cnb/src/model"
	"map-storage-cnb/src/utils"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

const (
	StorageTypeGitStorage model.StorageType = "GitStorage"
)

type FileObj struct {
	Name    string
	Hash    string
	Content []byte
}

type StorageFileMeta struct {
	Filename string
	Hash     string
	Status   model.MapStorageStatus
	Reason   string
}

type GitStorage struct {
	cfg                 model.GitStorageConfig
	DB                  *StorageDB
	ctx                 context.Context
	ctxCancel           context.CancelFunc
	wg                  sync.WaitGroup
	fileChan            chan FileObj
	StorageFileMetaChan chan StorageFileMeta
}

func NewGitStorage() *GitStorage {
	return &GitStorage{}
}

func (g *GitStorage) Init(cfg model.StorageConfig) error {
	g.cfg = cfg.GitStorage
	g.ctx, g.ctxCancel = context.WithCancel(context.Background())
	g.wg = sync.WaitGroup{}

	db, err := DBInit(cfg)
	if err != nil {
		g.ctxCancel()
		return err
	}
	g.DB = db

	g.DB = db
	g.fileChan = make(chan FileObj, g.cfg.MaxPushFileAtOnce*2)
	g.StorageFileMetaChan = make(chan StorageFileMeta, g.cfg.MaxPushFileAtOnce*2)

	g.initGitService()
	return nil
}

func (g *GitStorage) Close() error {
	g.ctxCancel()
	g.wg.Wait()
	close(g.fileChan)
	close(g.StorageFileMetaChan)
	return g.DB.Close()
}

func (g *GitStorage) Save(ctx context.Context, metaData model.MapMetaData, data []byte) (*model.MapMetaData, error) {
	metaData.SetStorageType(StorageTypeGitStorage)
	metaData.SetStorageStatus(model.MapUploadStatusOnProgress, "")
	g.DB.Add(ctx, metaData)
	g.fileChan <- FileObj{Name: metaData.Name, Hash: metaData.Hash, Content: data}
	return nil, nil
}

// TODO 未实现从git storage中读取文件的功能
func (g *GitStorage) Get(ctx context.Context, hash string, writer io.Writer) (*model.MapMetaData, error) {
	return nil, nil
}

func (g *GitStorage) GetMeta(ctx context.Context, hash string) (*model.MapMetaData, error) {
	return g.DB.Get(ctx, hash)
}

func (g *GitStorage) GetHistory(ctx context.Context, hash string, limit int) ([]model.MapMetaData, error) {
	var result []model.MapMetaData
	for {
		metaData, err := g.DB.Get(ctx, hash)
		if err != nil {
			return nil, err
		}
		result = append(result, *metaData)
		hash = metaData.PrevHash
		if hash == "" {
			break
		}
	}
	return result, nil
}

func (g *GitStorage) Exists(ctx context.Context, hash string) (bool, error) {
	_, err := g.DB.Get(ctx, hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (g *GitStorage) SearchExact(ctx context.Context, name string, limit int) ([]model.MapMetaData, error) {
	return g.DB.SearchExact(ctx, name, limit)
}

func (g *GitStorage) Search(ctx context.Context, name string, limit int) ([]model.MapMetaData, error) {
	return g.DB.Search(ctx, name, limit)
}

func (g *GitStorage) List(ctx context.Context, page int, desc bool, orderField string, limit int) ([]model.MapMetaData, error) {
	return g.DB.List(ctx, page, desc, orderField, limit)
}

// TODO 未实现git storage的文件删除功能
func (g *GitStorage) Delete(ctx context.Context, hash string) error {
	// TODO 删除git仓库内文件的功能还未实现

	err := g.DB.Delete(ctx, hash)
	if err != nil {
		return err
	}
	return nil

}

func cleanUp(dirPath string) {
	log.Printf("Clean up dir %q", dirPath)
	os.RemoveAll(dirPath)
}

func (g *GitStorage) initGitService() {

	var repo *git.Repository
	var err error

	if isGitInit(g.cfg.GitWorkSpaceDir) {
		log.Println("GitRepo is already init")
		repo, err = git.PlainOpen(g.cfg.GitWorkSpaceDir)
	} else {
		repo, err = g.gitClone(g.cfg.GitWorkSpaceDir)
		if errors.Is(err, transport.ErrEmptyRemoteRepository) {
			log.Println("Remote Repo is Empty, try to init")
			repo, err = g.gitInit(g.cfg.GitWorkSpaceDir)
		}
	}

	if err != nil {
		defer cleanUp(g.cfg.GitWorkSpaceDir)
		log.Fatalf("git prepare error : %v", err)
		return
	}

	workTree, err := repo.Worktree()
	if err != nil {
		log.Fatalf("Get Git Worktree Error : %v", err)
		return
	}
	if err = gitPull(workTree, ""); err != nil {
		log.Fatalf("Git Pull Error : %v", err)
		return
	}

	g.wg.Add(2)
	go func() {
		defer g.wg.Done()
		g.gitPushService(g.ctx, repo, workTree, g.cfg.GitWorkSpaceDir, g.fileChan)
	}()
	go func() {
		defer g.wg.Done()
		g.updateFileMetaToDB(g.ctx)
	}()

}

func (g *GitStorage) updateFileMetaToDB(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(g.StorageFileMetaChan)
			return
		case fileMeta := <-g.StorageFileMetaChan:
			var err error
			log.Printf("Update metadata record for %s", fileMeta.Filename)

			DBMeta := model.NewMetaData(fileMeta.Hash, fileMeta.Filename)
			DBMeta.StorageStatus = fileMeta.Status
			DBMeta.StorageStatusMsg = fileMeta.Reason

			err = g.DB.Update(ctx, DBMeta)
			if err != nil {
				log.Printf("failed to record failed metadata for %s: %v", DBMeta.Name, err)
			}

		}
	}
}

func setFileMetaMapAllFailed(storageFileMetaMap map[string]StorageFileMeta, reason string) {
	for hash := range storageFileMetaMap {
		fileMeta := storageFileMetaMap[hash]
		fileMeta.Status = model.MapUploadStatusFailed
		fileMeta.Reason = reason
		storageFileMetaMap[hash] = fileMeta
	}
}

func (g *GitStorage) gitPushService(ctx context.Context, repo *git.Repository, workTree *git.Worktree, dirPath string, fileChan chan FileObj) {
	eg, _ := errgroup.WithContext(ctx)
	eg.SetLimit(int(g.cfg.WriteFileWorkers))

	for {
		var err error
		storageFileMetaMap := make(map[string]StorageFileMeta)

		batch := g.collectFile(fileChan)
		if len(batch) == 0 {
			continue
		}
		batchFileNumber := len(batch)

		storageFiles := writeFileBatch(batch, dirPath, eg)
		for _, storageFile := range storageFiles {
			storageFileMetaMap[storageFile.Hash] = storageFile
		}

		log.Printf("Adding all file to git index")
		err = workTree.AddGlob("*")
		if err != nil {
			errStr := fmt.Sprintf("Git Add Error : %v", err)
			log.Println(errStr)
			setFileMetaMapAllFailed(storageFileMetaMap, errStr)
			continue
		}

		log.Printf("Creating commit")
		commitTitle := fmt.Sprintf("%d maps , %s", batchFileNumber, utils.ISO8601LocalNow())
		_, err = workTree.Commit(commitTitle, &git.CommitOptions{
			Author: &object.Signature{Name: g.cfg.CommitAuthor, Email: g.cfg.CommitEmail, When: time.Now()},
		})
		if err != nil {
			if errors.Is(err, git.ErrEmptyCommit) {
				log.Println("Git commit is empty, skip")
				continue
			}
			errStr := fmt.Sprintf("Git create commit Error : %v", err)
			log.Println(errStr)
			setFileMetaMapAllFailed(storageFileMetaMap, errStr)
			continue
		}

		err = g.gitPush(repo, "")
		if err != nil {
			errStr := fmt.Sprintf("Git Push Error : %v", err)
			log.Println(errStr)
			setFileMetaMapAllFailed(storageFileMetaMap, errStr)
			continue
		}
		for _, fileMeta := range storageFileMetaMap {
			g.StorageFileMetaChan <- fileMeta
		}
	}
}

// n秒窗口收一批
func (g *GitStorage) collectFile(fileChan <-chan FileObj) []FileObj {
	// log.Println("Collecting input files")
	var batch []FileObj

	waitSecond := time.Duration(g.cfg.FileBatchWindow) * time.Second
	timer := time.NewTimer(waitSecond)
	defer timer.Stop()

	for {
		select {
		case file, ok := <-fileChan:
			if !ok { // 通道关闭
				return batch
			}
			batch = append(batch, file)

			// 缓冲区满了 → 立即返回
			if len(batch) >= int(g.cfg.MaxPushFileAtOnce) {
				log.Printf("File batch buff are full")
				timer.Stop() // 1. 停止当前计时
				return batch
			}
		case <-timer.C:
			// log.Printf("File batch timeout in %d second", g.cfg.FileBatchWindow)
			return batch
		}
	}
}

// 用 errgroup 池化并发写文件
func writeFileBatch(batch []FileObj, dirPath string, eg *errgroup.Group) []StorageFileMeta {
	log.Printf("Writing %d files to %q", len(batch), dirPath)
	var mu sync.Mutex
	var result []StorageFileMeta

	for _, file := range batch {
		eg.Go(func() error {
			filePath := filepath.Join(dirPath,
				utils.AddSuffixIfMissing(file.Name, "map"))
			if err := os.WriteFile(filePath, file.Content, 0644); err != nil {
				mu.Lock()
				result = append(result, StorageFileMeta{
					Filename: file.Name,
					Hash:     file.Hash,
					Reason:   fmt.Sprintf("failed to write file %q: %v", file.Name, err),
					Status:   model.MapUploadStatusFailed,
				})
				mu.Unlock()
				return err // 依旧让 errgroup 能感知失败
			}
			mu.Lock()
			result = append(result, StorageFileMeta{
				Filename: file.Name,
				Hash:     file.Hash,
				Reason:   model.MapUploadStatusMsgSuccess,
				Status:   model.MapUploadStatusSuccess,
			})
			mu.Unlock()
			return nil
		})
	}
	_ = eg.Wait() // 不需要第一个错误，下面自己返回全部

	return result
}

// branch 传空值默认使用 "origin"
func gitPull(workTree *git.Worktree, branch string) error {
	if branch == "" {
		branch = "origin"
	}
	log.Printf("Pulling git repo in branch %q ", branch)
	err := workTree.Pull(&git.PullOptions{
		RemoteName: branch,
	})

	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		log.Println("Git Repo is already up to date")
		return nil
	}
	if err != nil {
		return err
	}
	log.Println("Git Repo pull successful")
	return nil
}
func isGitInit(gitDirPath string) bool {
	log.Printf("Checking %q git repo is initialized", gitDirPath)
	if !utils.DirExists(gitDirPath) {
		return false
	}
	_, err := git.PlainOpen(gitDirPath)
	return err == nil

}

func (g *GitStorage) gitClone(gitDirPath string) (*git.Repository, error) {
	log.Printf("Clone git repo %q to %q ", g.cfg.RemoteGitRepoUrl, gitDirPath)
	repo, err := git.PlainClone(gitDirPath, false, &git.CloneOptions{
		URL:      g.cfg.RemoteGitRepoUrl,
		Progress: os.Stdout,
	})
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (g *GitStorage) gitInit(gitDirPath string) (*git.Repository, error) {
	log.Printf("Init git repo for %q ", gitDirPath)

	cleanUp(g.cfg.GitWorkSpaceDir)

	repo, err := git.PlainInit(gitDirPath, false)
	if err != nil {
		return nil, err
	}

	// 2. 写一个初始文件并提交
	wt, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	readme := filepath.Join(gitDirPath, "README.md")
	err = os.WriteFile(readme, []byte("# hello go-git\n"), 0644)
	if err != nil {
		return nil, err
	}

	_, err = wt.Add("README.md")
	if err != nil {
		return nil, err
	}

	_, err = wt.Commit("initial repo", &git.CommitOptions{
		Author: &object.Signature{Name: g.cfg.CommitAuthor, Email: g.cfg.CommitEmail, When: time.Now()},
	})
	if err != nil {
		return nil, err
	}

	// 3. 添加远程
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{g.cfg.RemoteGitRepoUrl},
	})
	if err != nil {
		return nil, err
	}

	// 4. 推送到远程空仓库
	err = g.gitPush(repo, "")
	if err != nil {
		log.Printf("Git Push Error : %v", err)
		return nil, err
	}
	return repo, nil
}

// branch 传空值默认使用 "origin"
func (g *GitStorage) gitPush(repo *git.Repository, branch string) error {
	if branch == "" {
		branch = "origin"
	}
	log.Printf("Pushing commit to remote %q ", branch)
	err := repo.Push(&git.PushOptions{
		RemoteName: branch,
		RefSpecs:   []config.RefSpec{"refs/heads/master:refs/heads/master"},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		log.Printf("Push commit Error : %v", err)
		return err
	}
	log.Printf("Push commit Success")
	return nil
}

// // TickFiles 每隔 n 秒往 fileChan 塞一个 FileObj，直到 ctx 被取消
// func TickFiles(ctx context.Context, fileChan chan<- FileObj) {
// 	ticker := time.NewTicker(CreateFileTime)
// 	defer ticker.Stop()

// 	counter := 0
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			close(fileChan)
// 			return // 外部取消即退出
// 		case <-ticker.C:
// 			log.Printf("Creating test file %d", counter)
// 			counter++
// 			fileChan <- FileObj{
// 				Name:    fmt.Sprintf("file%d.txt", counter),
// 				Content: []byte(fmt.Sprintf("content-%d", counter)),
// 			}
// 		}
// 	}
// }
