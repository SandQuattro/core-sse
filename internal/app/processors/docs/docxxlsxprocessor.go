package docxxlsxprocessor

import (
	"bytes"
	"fmt"
	"github.com/nguyenthenguyen/docx"
	"github.com/thedatashed/xlsxreader"
	"io"
	"mime/multipart"
	"regexp"
	"strings"
)

func ProcessDocument(file *multipart.FileHeader) (string, error) {
	// Открываем файл
	doc, err := file.Open()
	if err != nil {
		return fmt.Sprintf("Ошибка открытия файла %s", file.Filename), err
	}
	defer doc.Close()
	size := file.Size

	data, err := io.ReadAll(doc)
	if err != nil {
		return fmt.Sprintf("Ошибка чтения содержимого файла %s", file.Filename), err
	}

	content, err := readDocx(data, size)
	if err == nil {
		return content, err
	}

	// если не смогли прочитать docx, пытаемся читать как xlsx
	content, err = readXlsx(data, 10)
	if err != nil {
		return "", err
	}

	return content, nil
}

func readDocx(data []byte, size int64) (string, error) {
	reader := bytes.NewReader(data)
	f, err := docx.ReadDocxFromMemory(reader, size)
	if err != nil {
		return "Ошибка чтения doc документа", err
	}
	defer f.Close()
	content := f.Editable().GetContent()

	// Remove XML tags
	re := regexp.MustCompile("<[^>]+>")
	text := re.ReplaceAllString(content, " ")

	// Чистый контент без тегов
	clean, err := strings.ReplaceAll(text, "  ", " "), nil
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(clean, "  ", " "), err
}

func readXlsx(bytes []byte, size int) (string, error) {
	// Create an instance of the reader by providing a data stream
	xl, _ := xlsxreader.NewReader(bytes)

	// Читаем первые 10 строк из файла
	records := make([][]string, 0, size)
	i := 0
	for row := range xl.ReadRows(xl.Sheets[0]) {
		i++
		cells := make([]string, 0, len(row.Cells))
		for _, cell := range row.Cells {
			cells = append(cells, cell.Value)
		}
		records = append(records, cells)

		if i >= size {
			break
		}
	}
	// Выводим первые 10 строк файла
	var text string
	for _, record := range records {
		text += strings.Join(record, ";") + "\n" // добавляем разделитель между полями и перенос строки между строками
	}

	return text, nil
}
