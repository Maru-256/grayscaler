package main

import (
	"flag"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io"
	"log"
	"os"
	"path/filepath"
)

func main() {
	boolFlag := flag.Bool("nobackup", false, "message for \"nobackup\"")
	flag.Parse()
	needBackup := !*boolFlag
	args := flag.Args()

	var files []string
	for _, arg := range args {
		fp, err := os.Open(arg)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		defer fp.Close()

		if isDir, _ := isDirectory(fp); isDir {
			err := filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				log.Println(err)
				continue
			}
		} else {
			files = append(files, arg)
		}
	}

	for _, arg := range files {
		fp, err := os.OpenFile(arg, os.O_RDWR, 0666)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		defer fp.Close()

		if needBackup {
			backup(fp, arg)
		}
		fp.Seek(0, os.SEEK_SET)

		img, _, err := image.Decode(fp)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		gray := toGray16(img)

		fp.Seek(0, os.SEEK_SET)
		if err := png.Encode(fp, gray); err != nil {
			log.Println(err.Error())
			continue
		}
	}
}

func backup(src *os.File, srcPath string) error {
	backupDir := filepath.FromSlash(filepath.Dir(srcPath) + "/" + "Originals/")
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		if err := os.Mkdir(backupDir, 0666); err != nil {
			return err
		}
	}

	dst, err := os.Create(filepath.FromSlash(backupDir) + "/" + filepath.Base(srcPath))
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}

func toGray16(src image.Image) *image.Gray16 {
	bounds := src.Bounds()
	dst := image.NewGray16(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			col := color.Gray16Model.Convert(src.At(x, y))
			dst.Set(x, y, col)
		}
	}
	return dst
}

func isDirectory(fp *os.File) (bool, error) {
	stat, err := fp.Stat()
	if err != nil {
		return false, err
	}
	return stat.IsDir(), nil
}
