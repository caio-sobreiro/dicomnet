package interfaces

import "github.com/caio-sobreiro/dicomnet/types"

// DataStore interface defines methods for persisting and retrieving DICOM data
type DataStore interface {
	// Patient operations
	FindPatients(query *types.QueryRequest) ([]types.Patient, error)
	GetPatient(patientID string) (*types.Patient, error)
	StorePatient(patient *types.Patient) error

	// Study operations
	FindStudies(query *types.QueryRequest) ([]types.Study, error)
	GetStudy(studyInstanceUID string) (*types.Study, error)
	StoreStudy(study *types.Study) error

	// Series operations
	FindSeries(query *types.QueryRequest) ([]types.Series, error)
	GetSeries(seriesInstanceUID string) (*types.Series, error)
	StoreSeries(series *types.Series) error

	// Image operations
	FindImages(query *types.QueryRequest) ([]types.Image, error)
	GetImage(sopInstanceUID string) (*types.Image, error)
	StoreImage(image *types.Image) error
}

// QueryProcessor interface for processing C-FIND queries
type QueryProcessor interface {
	ProcessQuery(query *types.QueryRequest) ([]interface{}, error)
}

// DatasetEncoder interface for encoding/decoding DICOM datasets
type DatasetEncoder interface {
	EncodeDataset(dataset *types.Dataset) []byte
	ParseDataset(data []byte) (*types.Dataset, error)
}
