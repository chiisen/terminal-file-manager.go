package app

// 狀態列永遠佔一列；小視窗依序省略通知與路徑列，保留可用的主要畫面。
func (m *AppState) viewRows() (headerRows, noticeRows, bodyHeight int) {
	if m.Height >= 3 {
		headerRows = 1
	}
	hasNotice := m.operationBusy || m.ErrorMessage != "" || (m.Mode == ModeNormal && m.StatusMessage != "" && m.StatusMessage != "newfile" && m.StatusMessage != "newdir")
	if hasNotice && m.Height >= 4 {
		noticeRows = 1
	}
	return headerRows, noticeRows, max(0, m.Height-1-headerRows-noticeRows)
}

func (m *AppState) previewSize() (width, height int) {
	_, _, height = m.viewRows()
	width = m.Width
	if width >= 70 && height >= 3 {
		width = width * 40 / 100
	}
	return width, height
}
