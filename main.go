// Copyright (c) 2025 Otto
// Лицензия: MIT (см. LICENSE)

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const CurrentVersion = "28.10.25" // Текущая версия FolderTree в формате "дд.мм.гг"

// Node структура представляет собой узел в дереве файловой системы
type Node struct {
	Name     string  // Cодержит имя файла или директории
	Path     string  // Cодержит полный путь до узла
	IsDir    bool    // Указывает, является ли узел директорией
	Children []*Node // Содержит список дочерних узлов
}

func main() {
	// Отображает версию FolderTree
	if len(os.Args) >= 2 && strings.EqualFold(os.Args[1], "--version") {
		fmt.Printf("Версия \"FolderTree\": %s\n", CurrentVersion)
		return
	}

	if len(os.Args) < 2 {
		// Выводит сообщение, потому что путь к директории является обязательным аргументом
		fmt.Println("Использование: укажите путь к директории как аргумент.")
		os.Exit(1)
	}

	dirPath := strings.Join(os.Args[1:], " ") // Собирает весь путь, даже если в нём есть пробелы

	info, err := os.Stat(dirPath)
	if err != nil {
		// Сообщает, если указанный путь недоступен
		fmt.Printf("Ошибка: невозможно получить доступ к указанному пути '%s': %v\n", dirPath, err)
		os.Exit(1)
	}

	if !info.IsDir() {
		// Сообщает, если указанный путь не является директорией
		fmt.Printf("Ошибка: '%s' — это файл, а не папка. Укажите путь к директории.\n", dirPath)
		os.Exit(1)
	}

	exePath, err := os.Executable()
	if err != nil {
		// Сообщает об ошибке, если невозможно определить путь исполняемого файла
		fmt.Printf("Ошибка получения пути исполняемого файла: %v\n", err)
		os.Exit(1)
	}
	exeDir := filepath.Dir(exePath)

	root, err := buildNode(dirPath)
	if err != nil {
		// Останавливает выполнение, если не удалось построить дерево из-за ошибки доступа или пути
		fmt.Printf("Ошибка построения дерева: %v\n", err)
		os.Exit(1)
	}

	// --- Рендеринг и запись файлов ---

	unicodeText := renderUnicodeTree(root)
	err = os.WriteFile(filepath.Join(exeDir, "Древо папок.txt"), []byte(unicodeText), 0644)
	if err != nil {
		// Игнорирует ошибку, если запись первого файла не удалась
		fmt.Printf("Ошибка записи 'Древо папок.txt': %v\n", err)
	}

	mdText := renderMarkdown(root)
	err = os.WriteFile(filepath.Join(exeDir, "Древо папок (Markdown).md"), []byte(mdText), 0644)
	if err != nil {
		// Игнорирует ошибку, если запись второго файла не удалась
		fmt.Printf("Ошибка записи 'Древо папок (Markdown).md': %v\n", err)
	}

	htmlText := renderHTML(root)
	err = os.WriteFile(filepath.Join(exeDir, "Древо папок (WEB).html"), []byte(htmlText), 0644)
	if err != nil {
		// Игнорирует ошибку, если запись третьего файла не удалась
		fmt.Printf("Ошибка записи 'Древо папок (WEB).html': %v\n", err)
	}

	// Сообщает пользователю, куда были сохранены результаты
	fmt.Printf("Созданы файлы по пути \"%s\":\n", exeDir)
	fmt.Println(" - Древо папок.txt")
	fmt.Println(" - Древо папок (Markdown).md")
	fmt.Println(" - Древо папок (WEB).html")
	//fmt.Println("Готово: 'Древо папок.txt', 'Древо папок (Markdown).md', 'Древо папок (WEB).html' созданы в", exeDir)
}

// buildNode рекурсивно строит структуру Node для заданного пути
func buildNode(path string) (*Node, error) {
	info, err := os.Stat(path)
	if err != nil {
		// Возвращает ошибку, если информация о пути недоступна
		return nil, err
	}
	node := &Node{
		Name:  info.Name(),
		Path:  path,
		IsDir: info.IsDir(),
	}
	if node.IsDir {
		entries, err := os.ReadDir(path)
		if err != nil {
			// Возвращает узел с ошибкой, если нет прав для чтения директории
			return node, err
		}
		sort.Slice(entries, func(i, j int) bool { return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name()) })
		// Сортирует по имени без учета регистра, чтобы вывод выглядел аккуратно
		for _, e := range entries {
			childPath := filepath.Join(path, e.Name())
			child, err := buildNode(childPath)
			if err != nil {
				// Пропускает узел, если возникла ошибка при доступе к нему (например, symlink loop или EPERM)
				continue
			}
			node.Children = append(node.Children, child)
		}
	}
	return node, nil
}

// renderUnicodeTree генерирует строковое представление дерева с использованием символов Unicode
func renderUnicodeTree(root *Node) string {
	var sb strings.Builder
	sb.WriteString(root.Name + "\n")
	renderUnicodeChildren(&sb, root, "")
	// Вызывает рекурсивную функцию для обработки дочерних элементов
	return sb.String()
}

// renderUnicodeChildren рекурсивно добавляет дочерние элементы с правильными префиксами Unicode
func renderUnicodeChildren(sb *strings.Builder, node *Node, prefix string) {
	for i, child := range node.Children {
		isLast := i == len(node.Children)-1
		if isLast {
			// Используется L-образный символ, потому что это последний элемент в списке
			sb.WriteString(prefix + "└── ")
		} else {
			// Используется T-образный символ, потому что за ним следуют другие элементы
			sb.WriteString(prefix + "├── ")
		}

		if child.IsDir {
			sb.WriteString(child.Name + "/\n")
		} else {
			sb.WriteString(child.Name + "\n")
		}

		if child.IsDir {
			if isLast {
				// Добавляет пробелы, чтобы ветка не продолжалась после последнего элемента
				renderUnicodeChildren(sb, child, prefix+"    ")
			} else {
				// Добавляет вертикальную линию, чтобы показать продолжение ветки
				renderUnicodeChildren(sb, child, prefix+"│   ")
			}
		}
	}
}

// renderMarkdown генерирует строковое представление дерева в формате Markdown
func renderMarkdown(root *Node) string {
	var sb strings.Builder
	//sb.WriteString("# Дерево: " + root.Name + "\n\n")
	renderMDNode(&sb, root, 0)
	return sb.String()
}

// renderMDNode рекурсивно строит элементы списка Markdown
func renderMDNode(sb *strings.Builder, n *Node, depth int) {
	indent := strings.Repeat("  ", depth)
	if depth == 0 {
		// Корневой элемент оформляется как заголовок или главный элемент
		sb.WriteString(indent + "📁 **" + n.Name + "**\n")
	} else {
		if n.IsDir {
			// Директории выделяются жирным шрифтом и символом папки
			sb.WriteString(indent + "- 📁 **" + n.Name + "**\n")
		} else {
			// Файлы используют символ листа и обычный текст
			sb.WriteString(indent + "- 📄 " + n.Name + "\n")
		}
	}
	for _, c := range n.Children {
		renderMDNode(sb, c, depth+1)
	}
}

// renderHTML генерирует полную HTML страницу, отображающую дерево
func renderHTML(root *Node) string {
	var sb strings.Builder
	// Записывает статический шаблон и стили
	sb.WriteString(`<!doctype html>
<html lang="ru">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Древо папок (WEB)</title>
<style>
body { font-family: Inter, system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial; padding: 18px; background:#f7f7fb; color:#111 }
.container { max-width: 1100px; margin: 0 auto; background: #fff; padding: 18px; border-radius: 10px; box-shadow: 0 6px 20px rgba(0,0,0,0.06); }
details { margin-left: 8px; }
summary { cursor: pointer; font-weight: 600; padding: 4px 0; }
.file { margin-left: 22px; padding: 2px 0; font-family: monospace; }
.meta { color:#666; font-size: 0.85em; margin-left:8px; }
.root { text-align: center; font-weight: 800; font-size: 1.35em; margin-bottom: 6px }
.small { font-size:0.9em; color:#666 }
</style>
</head>
<body>
<div class="container">
<div class="root">Структура папок (можно открывать и закрывать кликами)</div>
<hr/>
`)
	buildHTMLNode(&sb, root)
	// Завершает HTML структуру
	sb.WriteString(`
</div>
</body>
</html>
`)
	return sb.String()
}

// buildHTMLNode рекурсивно создает структуру HTML с использованием тегов details для директорий
func buildHTMLNode(sb *strings.Builder, n *Node) {
	for _, c := range n.Children {
		if c.IsDir {
			// Использует details/summary для создания раскрывающихся секций
			sb.WriteString("<details open>\n")
			sb.WriteString("<summary>📁 " + escapeHTML(c.Name) + "/</summary>\n")
			buildHTMLNode(sb, c)
			sb.WriteString("</details>\n")
		} else {
			// Файлы просто добавляются как div
			sb.WriteString("<div class=\"file\">📄 " + escapeHTML(c.Name) + "</div>\n")
		}
	}
}

// escapeHTML заменяет специальные символы HTML на их сущности, чтобы избежать проблем при рендеринге
func escapeHTML(s string) string {
	r := strings.ReplaceAll(s, "&", "&amp;")
	r = strings.ReplaceAll(r, "<", "&lt;")
	r = strings.ReplaceAll(r, ">", "&gt;")
	return r
}
