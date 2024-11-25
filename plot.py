import matplotlib.pyplot as plt
import csv
import os
import glob
import sys
import argparse
from collections import deque
from datetime import datetime


parser = argparse.ArgumentParser(description="处理命令行参数")

parser.add_argument('-data', default='/var/log/speed.log*', help='数据文件路径')
parser.add_argument('-line', type=int, default=120, help='行数')
parser.add_argument('-out', default='/var/www/html/speedtest.jpeg', help='输出文件路径')
parser.add_argument('-label', default='chisel', help='曲线标注')
parser.add_argument('-title', default='Speedtest Overview', help='图形标题')
parser.add_argument('-xlabel', default='Time', help='x轴标注')
parser.add_argument('-ylabel', default='Speed (Bps)', help='Y轴标注')
parser.add_argument('-xx', type=int, default=1000, help='数据倍率')

# 解析命令行参数
args = parser.parse_args()

# 访问参数值
log_files_pattern = args.data 
maxline = args.line
savefile= args.out

# 使用glob库获取符合模式的所有日志文件
log_files = sorted(glob.glob(log_files_pattern),key=os.path.getmtime, reverse=True)
# 存储读取的行
last_lines = deque(maxlen=maxline)

maxcol=0
# 读取日志文件，确保读取到最新的120行数据
for log_file in log_files:
    print("读取",log_file)
    with open(log_file, 'r') as file:
        lines = file.readlines()
        last_lines1=deque(maxlen=maxline)
        # 将当前文件的行追加到deque中
        for line in lines:
            line1=line.strip()
            row = line1.split(',')
            if maxcol < len(row):
                maxcol=len(row)
            last_lines1.append(line.strip())  # 去掉行尾的换行符
        for line in last_lines:
            last_lines1.append(line.strip())
        last_lines=last_lines1
        if len(last_lines) >= (maxline):
            break  # 如果已经获取了120行，停止读取

num_lines = len(last_lines)
print("行数",num_lines,"  列数",maxcol)
print("从",last_lines[0].split(',')[0],"到",last_lines[len(last_lines)-1].split(',')[0])
# 初始化数据列表
times = []
coldata=[]
for i in range(0,maxcol):
    data1=[]
    for line in last_lines:
        row = line.split(',')
        while len(row)<maxcol:
            row.append("0")
        try:
            datetime.strptime(row[0], '%Y-%m-%d %H:%M:%S')  # 调整时间格式根据你的文件
        except ValueError:
        # 如果转换失败，打印错误信息并跳过该行
            print(f"Skipping line due to error: {line}")
            continue  # 忽略有问题的行，继续处理下一行
        if i==0:
            times.append(datetime.strptime(row[0], '%Y-%m-%d %H:%M:%S'))
        else:
            try:
                f1=float(row[i])/args.xx
            except ValueError:
                f1=0.0
            data1.append(f1)
    if i>0:
        coldata.append(data1)



# 创建折线图
plt.figure(figsize=(20, 6))  # 图表大小
for i in range(0,maxcol-1):
    plt.plot(times, coldata[i], label=args.label+str(i))

# 添加图表标题和图例
plt.title(args.title)
plt.xlabel(args.xlabel)
plt.ylabel(args.ylabel)
plt.legend()

# 保存为 JPEG 文件
plt.savefig(savefile, format='jpeg')
plt.close()  # 关闭图表以释放内存

print("图表已保存为",savefile)
