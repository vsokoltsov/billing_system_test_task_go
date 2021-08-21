type IFileMarshaller interface {
	MarshallOperation(operation *WalletOperation) (*MarshalledResult, error)
	WriteToFile(mr *MarshalledResult) error
}

type JSONHandler struct {
	file *os.File
	mu   *sync.Mutex
}

func (jh *JSONHandler) MarshallOperation(operation *WalletOperation) (*MarshalledResult, error) {
	jsonBytes, jsonMarshallErr := json.Marshal(operation)
	if jsonMarshallErr != nil {
		return nil, fmt.Errorf("error of json marshalling: %s", jsonMarshallErr)
	}
	newLine := []byte("\n")
	data := append(jsonBytes, newLine...)
	return &MarshalledResult{
		id:   operation.ID,
		data: data,
	}, nil
}

func (jh *JSONHandler) WriteToFile(mr *MarshalledResult) error {
	bytesData := mr.data.([]byte)
	jh.mu.Lock()
	jh.file.Sync()
	jh.file.Write(bytesData)
	jh.file.Sync()
	jh.mu.Unlock()
	return nil
}

type CSVHandler struct {
	csvWriter *csv.Writer
	mu        *sync.Mutex
}

func (ch *CSVHandler) MarshallOperation(operation *WalletOperation) (*MarshalledResult, error) {
	idStr := strconv.Itoa(operation.ID)
	walletFromStr := strconv.Itoa(int(operation.WalletFrom.Int32))
	walletToStr := strconv.Itoa(int(operation.WalletTo))
	amountStr := operation.Amount.String()
	createdAtStr := operation.CreatedAt.String()
	row := []string{
		idStr,
		operation.Operation,
		walletFromStr,
		walletToStr,
		amountStr,
		createdAtStr,
	}
	return &MarshalledResult{
		id:   operation.ID,
		data: row,
	}, nil
}

func (ch *CSVHandler) WriteToFile(mr *MarshalledResult) error {
	row := mr.data.([]string)
	ch.mu.Lock()
	ch.csvWriter.Write(row)
	ch.mu.Unlock()
	return nil
}