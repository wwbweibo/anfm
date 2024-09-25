package tasks

import (
	"context"
	"fileserver/internal/adapters/dl"
	domainFile "fileserver/internal/domain/file"
	"fileserver/utils"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
)

type ScanOptions struct {
	RootPath   string
	Path       []string
	RegexPath  []*regexp.Regexp
	Extensions []string
}

func (opts ScanOptions) OptionRootPath(path string) ScanOptions {
	opts.RootPath = path
	return opts
}

func (opts ScanOptions) OptionPlainPath(path ...string) ScanOptions {
	opts.Path = append(opts.Path, path...)
	return opts
}

func (opts ScanOptions) OptionExtensions(ext ...string) ScanOptions {
	opts.Extensions = append(opts.Extensions, ext...)
	return opts
}

func (opts ScanOptions) OptionRegexPath(regex ...string) ScanOptions {
	for _, r := range regex {
		opts.RegexPath = append(opts.RegexPath, regexp.MustCompile(r))
	}
	return opts
}

func (opts ScanOptions) fileInPath(file string) bool {
	for _, p := range opts.Path {
		if strings.HasPrefix(file, p) {
			return true
		}
	}
	return false
}

func (opts ScanOptions) fileInRegexPath(file string) bool {
	for _, r := range opts.RegexPath {
		if r.MatchString(file) {
			return true
		}
	}
	return false
}

func (opts ScanOptions) fileInExtensions(file string) bool {
	for _, ext := range opts.Extensions {
		if strings.HasSuffix(file, ext) {
			return true
		}
	}
	return false
}

type SysInitBackendTask struct {
	startTime            time.Time
	option               ScanOptions
	repo                 domainFile.IFileRepository
	dlConfig             dl.Config
	imageCompressionTask ImageCompressionTask
}

func NewSysInitBackendTask(option ScanOptions,
	repo domainFile.IFileRepository,
	dlConfig dl.Config,
	imageCompressionTask ImageCompressionTask,
) *SysInitBackendTask {
	return &SysInitBackendTask{
		option:               option,
		repo:                 repo,
		dlConfig:             dlConfig,
		imageCompressionTask: imageCompressionTask,
	}
}

func (s *SysInitBackendTask) GetTaskName() string {
	return "sys_init_backend"
}

func (s *SysInitBackendTask) GetRunningDuration() time.Duration {
	return time.Duration(0)
}

func (s *SysInitBackendTask) Start(ctx context.Context) error {
	s.startTime = time.Now()
	files := utils.WalkDir(s.option.RootPath)
	log.Default().Printf("found %d files", len(files))
	for _, file := range files {
		select {
		case <-ctx.Done():
			break
		default:
			if s.option.fileInPath(file) || s.option.fileInRegexPath(file) || s.option.fileInExtensions(file) {
				domainFile.Root.Add(utils.GetDirectory(strings.ReplaceAll(file, s.option.RootPath, "")))
				s.singleFileHandler(ctx, file)
			}
		}
	}
	return nil
}

func (s *SysInitBackendTask) Stop(ctx context.Context) error {
	return nil
}

func (s *SysInitBackendTask) singleFileHandler(ctx context.Context, file string) {
	log.Default().Printf("handling file %s", file)
	file = strings.Replace(file, s.option.RootPath, "", 1)
	_file := domainFile.NewFile(file)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		result, err := dl.NewClient(s.dlConfig).Understanding(ctx, dl.UnderstandingRequest{
			Path: file,
		})
		if err != nil {
			log.Default().Printf("error getting file type: %v", err)
			return
		}
		_file.SetFileTypeFromUnderstanding(result)
	}()
	go func() {
		defer wg.Done()
		_file.Checksum = utils.Sha256(s.option.RootPath + file)
		_file.Size = utils.GetFileSize(s.option.RootPath + file)
	}()
	// insert into database
	wg.Wait()
	if _file.Group == "image" {
		s.imageCompressionTask.AddImage(_file)
	}
	err := s.repo.CreateOrUpdateFile(ctx, _file)
	if err != nil {
		log.Default().Printf("error inserting file %s: %v", file, err)
	}
}
