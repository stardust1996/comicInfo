package cbz

import (
	"archive/zip"
	"comicInfo/xml"
	"errors"
	"fmt"
	"github.com/pterm/pterm"
	"github.com/xuri/excelize/v2"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

/**
 * 2024/2/5
 * add by stardust
**/

// excelName 储存目录信息跟书籍信息的excel
const excelName = "info.xlsx"

// excel中字段序号
const (
	target = iota
	title
	summary
	web
	series
	writer
	penciller

	publisher
)

var statusMap = map[string]string{
	"已完结":  "ENDED",
	"连载中":  "ONGOING",
	"已放弃":  "ABANDONED",
	"有生之年": "HIATUS",
}

var directionMap = map[string]string{
	"从左往右":    "LEFT_TO_RIGHT",
	"从右往左":    "RIGHT_TO_LEFT",
	"垂直模式":    "VERTICAL",
	"Webtoon": "WEBTOON",
}

type dirNames struct {
	OldName     string           //旧目录名称
	SonDirs     []string         //旧子目录名称
	BookInfo    *xml.ComicInfo   //书籍信息
	ChapterInfo []*xml.ComicInfo //章节信息或者卷信息
}

func GetInfo() error {
	// 打开信息excel
	f, err := excelize.OpenFile(excelName)
	if err != nil {
		return errors.New("打开书籍及目录信息失败，请确认文件名")
	}

	rows, err := f.GetRows("书籍信息")
	if err != nil {
		return errors.New("获取书籍信息内容失败,失败原因" + err.Error())
	}
	nameList := make([]*dirNames, len(rows)-1)

	for i, row := range rows {
		if i == 0 {
			continue
		}

		if e := checkRequired(row, i, false); e != nil {
			return errors.New(fmt.Sprintf("第%d行书籍必填信息检查未通过，请检查信息:%s", i, e.Error()))
		}
		fmt.Printf("开始获取 %s 书籍信息\r\n", row[title])
		if row[target] == row[title] {
			return errors.New("因为需要在原地创建新文件夹,源文件夹和新文件夹名字不能相同，请检查信息")
		}
		info, err := os.Stat(row[target])
		if err != nil || !info.IsDir() {
			return errors.New(fmt.Sprintf("%s 不是一个有效路径或者文件夹", row[title]))
		}
		//初始化书籍信息
		nameList[i-1] = &dirNames{
			OldName: row[target],
			BookInfo: &xml.ComicInfo{
				Title:     row[title],
				Series:    row[series],
				Summary:   row[summary],
				Writer:    row[writer],
				Penciller: row[penciller],
				Web:       row[web],
				Publisher: row[publisher],
			},
		}

		fmt.Printf("开始获取 %s 章节信息\r\n", row[title])
		chapters, err := f.GetRows(row[title])
		if err != nil {
			return errors.New(fmt.Sprintf("%s 的章节信息获取失败", row[title]))
		}

		nameList[i-1].SonDirs = make([]string, len(chapters)-1)
		nameList[i-1].ChapterInfo = make([]*xml.ComicInfo, len(chapters)-1)

		// 进度条
		progressBar1, _ := pterm.DefaultProgressbar.WithTotal(len(chapters)).WithTitle(fmt.Sprintf("正在获取 %s 书籍信息...", row[title])).Start()

		for j, chapter := range chapters {
			progressBar1.Add(1)
			if j == 0 {
				continue
			}
			if e := checkRequired(chapter, i, true); e != nil {
				return errors.New(fmt.Sprintf("%s的第%d章节必填信息检查未通过，请检查信息：%s", row[title], i, e.Error()))
			}
			info, err = os.Stat(row[target] + "/" + chapter[target])
			if err != nil || !info.IsDir() {
				return errors.New(fmt.Sprintf("%s 不是一个有效路径或者文件夹:%s", chapter[target], err))
			}
			nameList[i-1].SonDirs[j-1] = chapter[target]
			chapterSummary, webUrl := "", ""
			if len(chapter) > summary {
				chapterSummary = chapter[summary]
			}
			if len(chapter) > web {
				webUrl = chapter[web]
			}
			nameList[i-1].ChapterInfo[j-1] = &xml.ComicInfo{
				Title:     chapter[title],
				Series:    row[series],
				Number:    strconv.Itoa(j),
				Summary:   chapterSummary,
				Writer:    row[writer],
				Penciller: row[penciller],
				Web:       webUrl,
				Publisher: row[publisher],
			}
			fmt.Printf("获取 %s 第 %d 章 %s 信息\r\n", row[title], j, chapter[title])
		}
		_, _ = progressBar1.Stop()
		fmt.Println(row[title] + "信息获取完毕")
	}

	// 执行操作
	for _, names := range nameList {
		//创建

		fmt.Println("创建漫画文件夹")
		err := os.Mkdir(names.BookInfo.Title, 0777)
		if err != nil {
			return errors.New(fmt.Sprintf("%s 创建文件夹失败:%s", names.BookInfo.Title, err))
		}

		//进度条
		// 进度条
		progressBar2, _ := pterm.DefaultProgressbar.WithTotal(len(names.ChapterInfo)).WithTitle(fmt.Sprintf("正在打包 %s...", names.BookInfo.Title)).Start()

		for i, info := range names.ChapterInfo {
			progressBar2.UpdateTitle(fmt.Sprintf("正在打包第%s话", getNo(i+1)))
			progressBar2.Add(1)
			oldPath := names.OldName + "/" + names.SonDirs[i]
			newPath := names.BookInfo.Title + "/" + info.Title + ".cbz"
			//创建xml文件
			xml.GenerateXML(oldPath, info)

			//章节打包
			err := compressDir(oldPath, newPath)
			if err != nil {
				return errors.New(fmt.Sprintf("%s 打包失败 %s", info.Title, err))
			}
		}
		_, _ = progressBar2.Stop()
		fmt.Println(names.BookInfo.Title + "打包完成")
		//生成书籍xml 暂时不支持书籍的,就不写这个了
		// xml.GenerateXML(names.BookInfo.Title, names.BookInfo)

		//整书打包
	}
	return nil
}

// dir: 需要打包的本地文件夹路径
// dst: 保存压缩包的本地路径
func compressDir(dir string, dst string) error {
	zipFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	defer func(zipFile *os.File) {
		_ = zipFile.Close()
	}(zipFile)
	archive := zip.NewWriter(zipFile)
	defer func(archive *zip.Writer) {
		_ = archive.Close()
	}(archive)
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// 如果是文件夹或者无法读取文件信息，则忽略
		if info.IsDir() || err != nil {
			return nil
		}

		// 打开文件
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func(file *os.File) {
			_ = file.Close()
		}(file)

		// 创建一个新的文件
		zipFile, err := archive.Create(info.Name())
		if err != nil {
			return err
		}

		// 将文件内容写入到 zip 文件中
		_, err = io.Copy(zipFile, file)
		if err != nil {
			return err
		}

		return nil
	})
}

func checkRequired(strList []string, i int, isChapter bool) error {
	if strings.TrimSpace(strList[target]) == "" {
		return errors.New("第" + strconv.Itoa(i+1) + "行目标文件夹信息为空")
	} else if strings.TrimSpace(strList[title]) == "" {
		return errors.New("第" + strconv.Itoa(i+1) + "行标题信息为空")
	} else if !isChapter && strings.TrimSpace(strList[series]) == "" {
		return errors.New("第" + strconv.Itoa(i+1) + "行系列信息为空")
	} else if !isChapter && strings.TrimSpace(strList[writer]) == "" {
		return errors.New("第" + strconv.Itoa(i+1) + "行书籍作者信息为空")
	} else if !isChapter && strings.TrimSpace(strList[publisher]) == "" {
		return errors.New("第" + strconv.Itoa(i+1) + "行书籍出版社信息为空")
	}
	return nil
}

func getNo(i int) string {
	s := "00" + strconv.Itoa(i)
	return s[len(s)-3:]
}
