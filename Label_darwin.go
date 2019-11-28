package zgo

func LabelNew(text string) *Label {
	return nil
}

func (l *Label) TextAlignment(a Alignment) View {
	l.alignment = a
	return l
}
