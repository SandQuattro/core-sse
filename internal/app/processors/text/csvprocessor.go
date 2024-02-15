package csvprocessor

import (
	"errors"
	logdoc "github.com/LogDoc-org/logdoc-go-appender/logrus"
	"github.com/jfyne/csvd"
	"io"
	"mime/multipart"
	"strings"
)

func ProcessCSVFile(file *multipart.FileHeader) (string, error) {
	logger := logdoc.GetLogger()
	logger.Debug("Processing text file ", file.Filename)

	// Открываем текстовый файл
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	sniffer := csvd.NewSniffer(15, ',', '\t', ';', ':', '|')
	reader := csvd.NewReader(src, sniffer)
	reader.LazyQuotes = true

	// Устанавливаем разделитель полей
	// reader.Comma = ';'

	// Читаем все строки из файла
	records, err := reader.ReadAll()
	if err != nil {
		return "", err
	}

	// Выводим содержимое файла
	var text string
	for _, record := range records {
		text += strings.Join(record, ";") + "\n" // добавляем разделитель между полями и перенос строки между строками
	}

	return text, nil
}

func ProcessCSVHeader(file *multipart.FileHeader, size int) (string, error) {
	logger := logdoc.GetLogger()
	logger.Debug("Processing csv header of the file ", file.Filename)

	// Открываем текстовый файл
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	sniffer := csvd.NewSniffer(15, ',', '\t', ';', ':', '|')
	reader := csvd.NewReader(src, sniffer)
	reader.LazyQuotes = true

	// Читаем первые n строк из файла, head + data
	records := make([][]string, 0, size)
	for i := 0; i < size; i++ {
		record, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}
		records = append(records, record)
	}

	// Выводим первые n строк файла, head + data
	var text string
	for _, record := range records {
		text += strings.Join(record, ";") + "\n" // добавляем разделитель между полями и перенос строки между строками
	}

	return text, nil
}
