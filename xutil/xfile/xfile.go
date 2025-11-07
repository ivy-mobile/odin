package xfile

import (
	"os"
	"path/filepath"
	"strings"
)

// IsExist checks if a file or directory exists at the given path.
//
// It takes a string parameter `path` which represents the path to the file or directory.
// It returns a boolean value indicating whether the file or directory exists or not.
func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

// JoinFilename 拼接filename
// 如:
// JoinFilename("/a/b/c.log", "-", "1")
// 返回: /a/b/c-1.log
func JoinFilename(filePath, sep, join string) string {
	// 分离目录和文件名
	dir := filepath.Dir(filePath)
	fileName := filepath.Base(filePath)

	// 处理文件名（不含扩展名）
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)

	newName := nameWithoutExt + sep + join

	// 重新组合路径
	return filepath.Join(dir, newName+ext)
}
