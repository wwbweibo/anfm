package file

import (
	"context"

	"gorm.io/gorm"
)

type IFileRepository interface {
	ListFileByDirectory(ctx context.Context, directory string) ([]File, error)
	GetFileByPath(ctx context.Context, path string) (File, error)
	CreateOrUpdateFile(ctx context.Context, file File) (err error)
	ListDirectory(ctx context.Context) ([]string, error)
}

type FileQuery struct {
	Path      string
	Directory string
	Extension string
	FileType  string
}

type FileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	db.AutoMigrate(&File{})
	return &FileRepository{
		db: db,
	}
}

func (r *FileRepository) ListDirectory(ctx context.Context) ([]string, error) {
	var dirs []string
	err := r.db.Model(&File{}).Select("distinct directory").Find(&dirs).Error
	if err != nil {
		return nil, err
	}
	return dirs, nil
}

func (r *FileRepository) GetFileByPath(ctx context.Context, path string) (File, error) {
	var file File
	err := r.db.Where("path = ?", path).First(&file).Error
	if err != nil {
		return File{}, err
	}
	return file, nil
}

func (r *FileRepository) ListFileByDirectory(ctx context.Context, directory string) ([]File, error) {
	// find directory under this directory
	var files []File
	err := r.db.Where("directory = ?", directory).Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (r *FileRepository) CreateOrUpdateFile(ctx context.Context, file File) (err error) {
	return r.db.Exec(`insert into files 
	(created_at, updated_at, path, directory, extension, type, "group", description, tags, caption, checksum, name, size ) values 
	(?,?,?,?,?,?,?,?,?,?,?, ?, ?) on conflict(path) do update set created_at = ?, updated_at = ?, directory = ?, extension = ?, type = ?, "group"=?, description=? , tags = ?,  caption = ?, checksum = ?, name=?, size=?;`,
		file.CreatedAt, file.UpdatedAt, file.Path, file.Directory, file.Extension, file.Type, file.Group, file.Description, file.Tags, file.Caption, file.Checksum, file.Name, file.Size,
		file.CreatedAt, file.UpdatedAt, file.Directory, file.Extension, file.Type, file.Group, file.Description, file.Tags, file.Caption, file.Checksum, file.Name, file.Size).Error
}
