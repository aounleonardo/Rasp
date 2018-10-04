package message

type Operation struct {
	Rumor *RumorRequest
	Identifier *IdentifierRequest
}

type RumorRequest struct {
	Contents string
}

type IdentifierRequest struct {}

type IdentifierResponse struct {
	Identifier string
}