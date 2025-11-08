package types

// QueryLevel represents the level of C-FIND query
type QueryLevel string

const (
	QueryLevelPatient QueryLevel = "PATIENT"
	QueryLevelStudy   QueryLevel = "STUDY"
	QueryLevelSeries  QueryLevel = "SERIES"
	QueryLevelImage   QueryLevel = "IMAGE"
)

// QueryRequest represents a parsed C-FIND query
type QueryRequest struct {
	Level              QueryLevel
	PatientName        string
	PatientID          string
	PatientBirthDate   string
	PatientSex         string
	StudyInstanceUID   string
	StudyID            string
	StudyDate          string
	StudyTime          string
	StudyDescription   string
	Modality           string
	SeriesInstanceUID  string
	SeriesNumber       string
	SeriesDescription  string
	SOPInstanceUID     string
	InstanceNumber     string
	AccessionNumber    string
	ReferringPhysician string
}

// Patient represents patient data
type Patient struct {
	Name      string
	ID        string
	BirthDate string
	Sex       string
	Studies   []Study
}

// Study represents study data
type Study struct {
	InstanceUID  string
	ID           string
	Date         string
	Time         string
	Description  string
	AccessionNum string
	RefPhysician string
	Series       []Series
}

// Series represents series data
type Series struct {
	InstanceUID string
	Number      string
	Description string
	Modality    string
	Images      []Image
}

// Image represents image data
type Image struct {
	SOPInstanceUID string
	InstanceNumber string
}
