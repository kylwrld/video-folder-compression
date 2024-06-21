package main

import (
	"flag"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var logger *log.Logger = log.New(os.Stdout, "[LOG] ", log.Ltime)

// Compress a single file
func Compress(input string, output string, file_name string) {
	// ffmpeg -i <input.mp4> -vcodec h264_nvenc -crf 10 -acodec aac -b:a 128k -b:v 5000k <output.mp4>

	input = input + "\\" + file_name
	output = output + "\\" + file_name
	
	logger.Println("| Starting compression:", file_name)

	cmd := exec.Command("ffmpeg", "-y", "-i",
		input,
		"-vcodec", "h264_nvenc", "-crf", "10", "-acodec", "aac", "-b:a", "128k", "-b:v", "5000k",
		output)

	if err := cmd.Run(); err != nil {
		logger.Println("| Unable to compress file:", file_name)
	}
}

// Compress a list of files
func CompressList(files *[]fs.DirEntry, input string, output string, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, file := range *files {
		Compress(input, output, file.Name())
	}
}

// Check if the number of files is greater than the number of goroutines.
func CheckGoroutines(goroutines_qtd int, qtd_files int) int {
	if qtd_files < goroutines_qtd {
		new_qtd := qtd_files
		logger.Println("| Unable to use", goroutines_qtd, "goroutines. The number has been reduced to", new_qtd)
		return new_qtd
	} 
	return goroutines_qtd
}

func SendGoroutines(goroutines_qtd int, qtd_files int, files *[]fs.DirEntry, input string, output string) {
	var wg sync.WaitGroup

	goroutines_qtd = CheckGoroutines(goroutines_qtd, qtd_files)
	quotient := qtd_files / goroutines_qtd
	start := 0
	count := quotient

	for i := 0; i < goroutines_qtd; i++ {
		if i == goroutines_qtd-1 {
			wg.Add(1)

			fls := (*files)[start:]
			go CompressList(&fls, input, output, &wg)
		} else {
			wg.Add(1)

			fls := (*files)[start:count]
			go CompressList(&fls, input, output, &wg)
		}
		start += quotient
		count += quotient
	}

	wg.Wait()
}

func DirSizeGB(path string) int64 {
    var dir_size int64 = 0

    read_size := func(path string, file os.FileInfo, err error) error {
        if !file.IsDir() {
            dir_size += file.Size()
        }
        return nil
    }

    filepath.Walk(path, read_size)    

    size_mb := dir_size / 1024 / 1024

    return size_mb
}

func main() {
	var input_path string
	var output_path string
	flag.StringVar(&input_path, "input", "", "input directory")
	flag.StringVar(&output_path, "output", "", "output directory")
	
	flag.Parse()
	
	goroutines_qtd := 5 
	files, err := os.ReadDir(input_path)
	qtd_files := len(files)
	size_mb := DirSizeGB(input_path)
	if err != nil {
		log.Fatal(err)
	}

	logger.Println("| Number of goroutines:", goroutines_qtd)
	logger.Println("| Starting to compress folder", input_path)
	logger.Println("| Number of files:", qtd_files)
	logger.Println("| Size MB:", size_mb, output_path)

	SendGoroutines(goroutines_qtd, qtd_files, &files, input_path, output_path)
	logger.Println("| Done.")
}
