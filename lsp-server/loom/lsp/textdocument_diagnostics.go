package lsp

type DocumentDiagnosticRequest struct {
	Request
	Params DocumentDiagnosticParams `json:"params"`
}

type DocumentDiagnosticParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type RelatedFullDocumentDiagnosticReport struct {
	RelatedDocuments map[string]FullDocumentDiagnosticReport `json:"relatedDocuments"`
}

func NewDiagnosticResponseHandler() *RelatedFullDocumentDiagnosticReport {
	return &RelatedFullDocumentDiagnosticReport{
		RelatedDocuments: make(map[string]FullDocumentDiagnosticReport),
	}
}

func (r *RelatedFullDocumentDiagnosticReport) AddDocument(uri string, report FullDocumentDiagnosticReport) {
	r.RelatedDocuments[uri] = report
}

type DiagnosticResponse struct {
	Response
	Result RelatedFullDocumentDiagnosticReport `json:"result"`
}

func NewDiagnosticResponse(id int, fullReport RelatedFullDocumentDiagnosticReport) DiagnosticResponse {
	return DiagnosticResponse{
		Response: Response{
			RPC: "2.0",
			ID:  &id,
		},
		Result: fullReport,
	}
}

type FullDocumentDiagnosticReport struct {
	Kind  DocumentDiagnosticKind `json:"kind"`
	Items []Diagnostic           `json:"items"`
}

type DocumentDiagnosticKind string // "full"

type Diagnostic struct {
	Range   Range  `json:"range"`
	Message string `json:"message"`
}


