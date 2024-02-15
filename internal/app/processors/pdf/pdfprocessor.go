package pdfprocessor

import (
	logdoc "github.com/LogDoc-org/logdoc-go-appender/logrus"
	"github.com/rudolfoborges/pdf2go"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func ProcessPdfFile(file *multipart.FileHeader) (string, error) {
	logger := logdoc.GetLogger()
	logger.Debug("Processing pdf file ", file.Filename)

	// Открываем файл
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Создать временный файл
	tempFile, err := os.Create(os.TempDir() + file.Filename) //nolint:gosec
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFile.Name())

	// Записать содержимое загруженного файла во временный файл
	_, err = io.Copy(tempFile, src)
	if err != nil {
		return "", err
	}

	pdfPath, err := filepath.Abs(tempFile.Name())
	if err != nil {
		return "", err
	}

	pdf, err := pdf2go.New(pdfPath, pdf2go.Config{
		LogLevel: pdf2go.LogLevelError,
	})

	if err != nil {
		return "", err
	}

	text, err := pdf.Text()
	if err != nil {
		return "", err
	}

	return text, nil

	// Если у нас бесплатный план, то выгружаем ограниченное количество страниц
	// pages, err := pdf.Pages()
	//
	// if err != nil {
	//	panic(err)
	//}
	//
	// for _, page := range pages {
	//	fmt.Println(page.Text())
	//}
	//
	// return "", nil
}
