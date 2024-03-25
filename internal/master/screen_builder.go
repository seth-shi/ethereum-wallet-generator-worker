package master

import (
	"bytes"
	"github.com/olekukonko/tablewriter"
	"net/url"
)

type ScreenBuilder struct {
	Table           *tablewriter.Table
	NeedClearScreen bool
	content         string
	buf             *bytes.Buffer
}

func newScreenBuilder() *ScreenBuilder {
	buf := &bytes.Buffer{}
	tab := tablewriter.NewWriter(buf)
	tab.SetHeader([]string{"#", "节点", "已找到", "已生成", "占比", "速度", "运行时间", "版本号"})

	return &ScreenBuilder{
		buf:             buf,
		Table:           tab,
		NeedClearScreen: true,
	}
}

func (o *ScreenBuilder) Build(data [][]string, footer []string) {

	o.Table.AppendBulk(data)
	o.Table.SetFooter(footer)
	o.Table.SetFooterAlignment(tablewriter.ALIGN_LEFT)
	o.Table.SetAlignment(tablewriter.ALIGN_LEFT)
	o.Table.Render()

	o.content = o.buf.String()
	o.Table.ClearRows()
	o.Table.ClearFooter()
	o.buf.Reset()
}

func (o *ScreenBuilder) GetContent() string {
	return o.content
}

func (o *ScreenBuilder) GetEncodeContent() string {
	return url.QueryEscape(o.GetContent())
}
