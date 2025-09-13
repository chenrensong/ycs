package contracts

// IContentFactory represents a factory interface for creating content objects without circular dependencies
type IContentFactory interface {
	// CreateContentType creates ContentType for AbstractType instances
	CreateContentType(abstractType IAbstractType) IContent
	// CreateContentDoc creates ContentDoc for YDoc instances
	CreateContentDoc(doc interface{}) IContent
	// CreateContentBinary creates content for binary data
	CreateContentBinary(ba []byte) IContent
	// CreateContentAny creates content for any value
	CreateContentAny(value interface{}) IContent
	// CreateContent creates appropriate content based on value type
	CreateContent(value interface{}) IContent
	// CreateContentFormat creates ContentFormat for formatting
	CreateContentFormat(key string, value interface{}) IContent
	// CreateContentString creates ContentString for string content
	CreateContentString(text string) IContent
	// CreateContentEmbed creates ContentEmbed for embedded content
	CreateContentEmbed(embed interface{}) IContent
}

// IContentReaderRegistry represents a registry for content readers to avoid circular dependencies
type IContentReaderRegistry interface {
	// RegisterContentReader registers a content reader for a specific content reference ID
	RegisterContentReader(contentRef int, reader func(IUpdateDecoder) IContent)
	// ReadContent reads content by content reference ID
	ReadContent(contentRef int, decoder IUpdateDecoder) IContent
}
