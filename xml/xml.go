package xml

import (
	"comicInfo/log"
	"encoding/xml"
	"os"
)

const xmlUrl1 = "http://www.w3.org/2001/XMLSchema"
const xmlUrl2 = xmlUrl1 + "-instance"
const xmlName = "/ComicInfo.xml"

/**
 * 2024/2/5
 * add by stardust
**/

type ComicInfo struct {
	XMLName   xml.Name `xml:"ComicInfo"`
	XSD       string   `xml:"xmlns:xsd,attr"`
	XSI       string   `xml:"xmlns:xsi,attr"`
	Title     string   `xml:"Title"`               //标题
	Series    string   `xml:"Series"`              //系列
	Number    string   `xml:"Number,omitempty"`    //序号
	Summary   string   `xml:"Summary,omitempty"`   //简介
	Writer    string   `xml:"Writer"`              //作者
	Penciller string   `xml:"Penciller,omitempty"` //线稿师
	Web       string   `xml:"Web,omitempty"`       //链接
	//Status           string   `xml:"Status"`              //状态
	Publisher string `xml:"Publisher,omitempty"` //出版社
	//ReadingDirection string   `xml:"ReadingDirection"`    //阅读方向
}

// GenerateXML
/*
 * 生成xml文件
 * path 生成文件路径
 * info 生成文件内容
 */
func GenerateXML(path string, info *ComicInfo) {
	//设置行内属性
	info.XSI = xmlUrl2
	info.XSD = xmlUrl1
	path += xmlName
	output, err := xml.MarshalIndent(info, " ", " ")
	if err != nil {
		log.Logger.Println("生成xml内容失败", err)
		return
	}
	file, err := os.Create(path)
	if err != nil {
		log.Logger.Println("生成xml文件失败", err)
		return
	}
	_, err = file.Write([]byte(xml.Header))
	if err != nil {
		log.Logger.Println("写入xml头信息失败", err)
		return
	}
	_, err = file.Write(output)
	if err != nil {
		log.Logger.Println("写入xml信息失败", err)
		return
	}
}
