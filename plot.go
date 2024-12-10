package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"image/color"

)
type File struct {
	Name    string
	ModTime time.Time
}

// sortFilesByModTime 接收文件路径模式，返回排序后的文件列表
func sortFilesByModTime(pattern string) ([]string, error) {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var fileDetails []File
	for _, fileName := range files {
		info, err := os.Stat(fileName)
		if err != nil {
			return nil, fmt.Errorf("error getting info for file %s: %v", fileName, err)
		}
		fileDetails = append(fileDetails, File{Name: fileName, ModTime: info.ModTime()})
	}

	// 按修改时间排序，最新的文件排在前面
	sort.Slice(fileDetails, func(i, j int) bool {
		return fileDetails[i].ModTime.After(fileDetails[j].ModTime)
	})
	files=[]string{}
	for _,file := range fileDetails {
		files = append(files,file.Name)
	}

	return files, nil
}
func customUsage() {
	fmt.Fprintf(os.Stderr, "简单时序折线图绘图工具，从文本文件读取逗号分隔的数据行，第一列为日期时间(2024-01-01 0:00:00)，后面若干列为数据\n")
	fmt.Fprintf(os.Stderr, "用法：%s [参数...]\n", os.Args[0])
	flag.PrintDefaults()  // 输出默认的帮助信息
	fmt.Fprintf(os.Stderr, "示例:\n  %s -x 1000 -ylabel \"Speed (KBps)\" -lines 200 -in d:\\test1.log -out d:\\test1.png\n", os.Args[0])
}

func main() {
	flag.Usage = customUsage
	filePattern := flag.String("in", "/var/log/speed.log*", "输入的数据文件，支持通配符，例如 /var/log/speed.log*")
	outfile := flag.String("out", "/var/www/html/speedtest.jpeg", "输出JPEG文件，也可以是PNG文件，例如 /var/www/html/speedtest.jpeg")
	title := flag.String("title", "Speedtest Overview", "图片标题")
	label := flag.String("label", "data-", "数据图线标注")
	xLabel := flag.String("xlabel", "Time", "横坐标标签")
	yLabel := flag.String("ylabel", "Speed (Bps)", "纵坐标标签")
	maxlines := flag.Int("lines", 120, "读取行数，从文件末尾开始算，如果0的话读取整个文件")
	width := flag.Int("width", 42, "图片宽度，厘米")
	height := flag.Int("height", 16, "图片高度，厘米")
	xx := flag.Float64("x", 1, "纵坐标除以数据倍率")
	avrg := flag.Bool("avrg", false, "绘制平均趋势虚线")
	
	flag.Parse()

	// 获取匹配的文件列表
	files, err := sortFilesByModTime(*filePattern)
	if err != nil {
		fmt.Println("获取文件出错:", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("无输入文件:", *filePattern)
		return
	}

	//fmt.Println(files)
	// 读取最后n行数据
	var allData [][]string
	linesRead := 0
	for _, file := range files {
		if linesRead >= *maxlines && *maxlines!=0 {
			break
		}
		fmt.Println("读取",file)
		fileData, err := readLastNLines(file, *maxlines-linesRead)
		if err != nil {
			fmt.Println("读取文件出错:", err)
			return
		}
		allData = append(fileData,allData...)
		linesRead += len(fileData)
	}

	maxColumns := findMaxColumns(allData)
	fmt.Println("读取行数：", len(allData),"  读取最大列数：", maxColumns)

	for i := range allData {
		for len(allData[i]) < maxColumns {
			allData[i] = append(allData[i], "0")
		}
	}

	err = plotData(allData, *xx, *title, *xLabel, *yLabel,*label,*width,*height,*outfile,*avrg)
	if err != nil {
		fmt.Println("绘图出错:", err)
		return
	}
	fmt.Println("图表已保存为",*outfile)
}

func readLastNLines(filename string, n int) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines [][]string
	scanner := bufio.NewScanner(file)

	// 将文件内容按行扫描并保存
	var allLines []string
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	// 取最后n行数据
	start := 0
	if n>0 {
	    start = len(allLines) - n
	}
	if start < 0 {
		start = 0
	}
	for _, line := range allLines[start:] {
		// 假设每行是逗号分隔的
		lines = append(lines, strings.Split(line, ","))
	}

	return lines, scanner.Err()
}

func findMaxColumns(data [][]string) int {
	maxColumns := 0
	for _, row := range data {
		if len(row) > maxColumns {
			maxColumns = len(row)
		}
	}
	return maxColumns
}

func plotData(data [][]string,xx float64, title, xLabel, yLabel string, label string, width,height int, jpgfile string, avrg bool ) error {

    lineColors := []color.Color{
        color.RGBA{R: 255, G: 0, B: 0, A: 255},     // 红色
        color.RGBA{R: 0, G: 255, B: 0, A: 255},     // 绿色
        color.RGBA{R: 0, G: 0, B: 255, A: 255},     // 蓝色
        color.RGBA{R: 0, G: 255, B: 255, A: 255},   // 青色
        color.RGBA{R: 255, G: 0, B: 255, A: 255},   // 品红
        color.RGBA{R: 128, G: 0, B: 0, A: 255},     // 暗红
        color.RGBA{R: 128, G: 128, B: 0, A: 255},   // 橄榄
        color.RGBA{R: 0, G: 128, B: 0, A: 255},     // 暗绿
        color.RGBA{R: 128, G: 0, B: 128, A: 255},   // 紫色
        color.RGBA{R: 0, G: 128, B: 128, A: 255},   // 暗青
        color.RGBA{R: 0, G: 0, B: 128, A: 255},     // 暗蓝
        color.RGBA{R: 255, G: 165, B: 0, A: 255},   // 橙色
        color.RGBA{R: 100, G: 100, B: 100, A: 255}, // 暗银色
        color.RGBA{R: 128, G: 128, B: 0, A: 255},   // 暗黄色
	}

	// 转换数据为图表需要的格式
	// 假设第一列为时间，后面的列为数据
	timeStamps := []float64{}
	columnsData := make([][]float64, len(data[0])-1)

	for _, row := range data {
		// 解析时间
		timestamp, err := time.Parse("2006-01-02 15:04:05", row[0])
		if err != nil {
		    fmt.Fprintf(os.Stderr,"%v, 日期格式错: %v\n",row, err)
		    continue
		}
		//fmt.Println(row)
		timeStamps =append(timeStamps,float64(timestamp.Unix()))

		// 处理后面的数据列
		for j := 1; j < len(row); j++ {
			val, err := strconv.ParseFloat(row[j], 64)
			if err != nil {
			    val = 0
				fmt.Fprintf(os.Stderr,"%v, 数据格式错: %v\n",row, err)
			}
			columnsData[j-1] = append(columnsData[j-1], val/xx)
		}
	}
	//fmt.Println(len(columnsData),len(columnsData[0]),columnsData)
	//fmt.Println(len(timeStamps),timeStamps)
	fmt.Println("有效数据行数：", len(timeStamps))
	fmt.Println("从",data[0][0],"到",data[len(data)-1][0])

	// 创建一个新的图形
	p:= plot.New()
	if  p== nil {
		return nil
	}

	p.Title.Text = title
	p.Title.Padding = vg.Length(1)*vg.Centimeter
	p.Title.TextStyle.Font.Size = vg.Points(28)
	p.X.Label.TextStyle.Font.Size = vg.Points(18)
	p.X.Label.Text = xLabel
	p.Y.Label.TextStyle.Font.Size = vg.Points(18)
	p.Y.Label.Text = yLabel
	p.Y.Label.Padding = vg.Length(2)*vg.Centimeter
	//p.Y.Padding= vg.Length(2)*vg.Centimeter

	// 解析数据并添加到图表
	for i := 0; i < len(columnsData); i++ {
		pts := make(plotter.XYs, len(columnsData[i]))
		apts := make(plotter.XYs, len(columnsData[i]))
		sum:=0.0
		for j, value := range columnsData[i] {
			pts[j].X = timeStamps[j]
			pts[j].Y = value
			sum=sum+value
			apts[j].X = timeStamps[j]
			apts[j].Y = sum/float64(j)
		}
	
		l, err := plotter.NewLine(pts)
		al, _ :=plotter.NewLine(apts)
		if err != nil {
			fmt.Sprintln("绘图出错: %v", err)
		}
		l.LineStyle.Width = vg.Points(1)
		l.LineStyle.Color = lineColors[i%len(lineColors)]
		al.LineStyle.Color = lineColors[i%len(lineColors)]
		al.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}
		if avrg {
			p.Add(al)
		}
		p.Add(l)
		p.Legend.Add(fmt.Sprintf("%s %d",label, i+1), l)
		
	}

	// 设置X轴为时间格式
	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02"}

	// 设置图例位置
	//p.Legend.Top = true
	p.Legend.Left = true
	p.Legend.Padding = vg.Length(3)
	p.Legend.YOffs = vg.Length(10)	
	p.Legend.XOffs = vg.Length(-4)*vg.Centimeter
	fmt.Println("生成文件",jpgfile)
	err := p.Save(vg.Length(width)*vg.Centimeter, vg.Length(height)*vg.Centimeter, jpgfile)
	
	// 保存为JPEG文件
	return err
}
