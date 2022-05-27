package packageInfo

import (
	"database/sql/driver"
	"github.com/pennsieve/pennsieve-go-api/models/fileInfo"
	"github.com/pennsieve/pennsieve-go-api/models/iconInfo"
)

type PackageController struct{}

// PackageState is an enum indicating the state of the Package
type State int64

const (
	Unavailable State = iota
	Uploaded
	Deleting
	Infected
	UploadFailed
	Processing
	Ready
	ProcessingFailed
)

func (s State) String() string {
	switch s {
	case Unavailable:
		return "UNAVAILABLE"
	case Uploaded:
		return "UPLOADED"
	case Deleting:
		return "DELETING"
	case Infected:
		return "INFECTED"
	case UploadFailed:
		return "UPLOAD_FAILED"
	case Processing:
		return "PROCESSING"
	case Ready:
		return "READY"
	case ProcessingFailed:
		return "PROCESSING_FAILED"
	}
	return "UNKNOWN"
}

func (s State) DBMap(value string) State {
	switch value {
	case "UNAVAILABLE":
		return Unavailable
	case "UPLOADED":
		return Uploaded
	case "DELETING":
		return Deleting
	case "INFECTED":
		return Infected
	case "UPLOAD_FAILED":
		return UploadFailed
	case "PROCESSING":
		return Processing
	case "READY":
		return Ready
	case "PROCESSING_FAILED":
		return ProcessingFailed
	}
	return Unavailable
}

func (u *State) Scan(value interface{}) error { *u = u.DBMap(value.(string)); return nil }
func (u State) Value() (driver.Value, error)  { return u.String(), nil }

// PackageType is an enum indicating the type of the Package
type Type int64

const (
	Image Type = iota
	MRI
	Slide
	ExternalFile
	MSWord
	PDF
	CSV
	Tabular
	TimeSeries
	Video
	Unknown
	Collection
	Text
	Unsupported
	HDF5
	ZIP
)

func (s Type) String() string {
	switch s {
	case Image:
		return "Image"
	case MRI:
		return "MRI"
	case Slide:
		return "Slide"
	case ExternalFile:
		return "ExternalFile"
	case MSWord:
		return "MSWord"
	case PDF:
		return "PDF"
	case CSV:
		return "CSV"
	case Tabular:
		return "Tabular"
	case TimeSeries:
		return "TimeSeries"
	case Video:
		return "Video"
	case Unknown:
		return "Unknown"
	case Collection:
		return "Collection"
	case Text:
		return "Text"
	case Unsupported:
		return "Unsupported"
	case HDF5:
		return "HDF5"
	case ZIP:
		return "ZIP"
	}
	return "Unknown"
}

func (s Type) DBMap(value string) Type {
	switch value {
	case "Image":
		return Image
	case "MRI":
		return MRI
	case "Slide":
		return Slide
	case "ExternalFile":
		return ExternalFile
	case "MSWord":
		return MSWord
	case "PDF":
		return PDF
	case "CSV":
		return CSV
	case "Tabular":
		return Tabular
	case "TimeSeries":
		return TimeSeries
	case "Video":
		return Video
	case "Unknown":
		return Unknown
	case "Collection":
		return Collection
	case "Text":
		return Text
	case "Unsupported":
		return Unsupported
	case "HDF5":
		return HDF5
	case "ZIP":
		return ZIP
	}
	return Unknown
}

func (u *Type) Scan(value interface{}) error { *u = u.DBMap(value.(string)); return nil }
func (u Type) Value() (driver.Value, error)  { return u.String(), nil }

type FileTypeInfo struct {
	PackageType    Type
	PackageSubType string
	Icon           iconInfo.Icon
}

// fileTypeDict maps filetypes to PackageTypes.
var FileTypeDict = map[fileInfo.Type]FileTypeInfo{
	fileInfo.MEF: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.EDF: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.TDMS: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.OpenEphys: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.Persyst: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.NeuroExplorer: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.MobergSeries: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.BFTS: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.Nicolet: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.MEF3: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.Feather: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.NEV: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.Spike2: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.MINC: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileInfo.DICOM: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileInfo.NIFTI: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileInfo.ROI: {
		PackageType:    Unsupported,
		PackageSubType: "Morphology",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileInfo.SWC: {
		PackageType:    Unsupported,
		PackageSubType: "Morphology",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileInfo.ANALYZE: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileInfo.MGH: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileInfo.JPEG: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           iconInfo.Image,
	},
	fileInfo.PNG: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           iconInfo.Image,
	},
	fileInfo.TIFF: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileInfo.OMETIFF: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileInfo.BRUKERTIFF: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileInfo.CZI: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileInfo.JPEG2000: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileInfo.LSM: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileInfo.NDPI: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileInfo.OIB: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileInfo.OIF: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileInfo.GIF: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           iconInfo.Image,
	},
	fileInfo.WEBM: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           iconInfo.Video,
	},
	fileInfo.MOV: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           iconInfo.Video,
	},
	fileInfo.AVI: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           iconInfo.Video,
	},
	fileInfo.MP4: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           iconInfo.Video,
	},
	fileInfo.CSV: {
		PackageType:    CSV,
		PackageSubType: "Tabular",
		Icon:           iconInfo.Tabular,
	},
	fileInfo.TSV: {
		PackageType:    CSV,
		PackageSubType: "Tabular",
		Icon:           iconInfo.Tabular,
	},
	fileInfo.MSExcel: {
		PackageType:    Unsupported,
		PackageSubType: "MS Excel",
		Icon:           iconInfo.Excel,
	},
	fileInfo.Aperio: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileInfo.MSWord: {
		PackageType:    MSWord,
		PackageSubType: "MS Word",
		Icon:           iconInfo.Word,
	},
	fileInfo.PDF: {
		PackageType:    PDF,
		PackageSubType: "PDF",
		Icon:           iconInfo.PDF,
	},
	fileInfo.Text: {
		PackageType:    Text,
		PackageSubType: "Text",
		Icon:           iconInfo.Text,
	},
	fileInfo.BFANNOT: {
		PackageType:    Unknown,
		PackageSubType: "Text",
		Icon:           iconInfo.Generic,
	},
	fileInfo.AdobeIllustrator: {
		PackageType:    Unsupported,
		PackageSubType: "Illustrator",
		Icon:           iconInfo.AdobeIllustrator,
	},
	fileInfo.AFNI: {
		PackageType:    Unsupported,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileInfo.AFNIBRIK: {
		PackageType:    Unsupported,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileInfo.Ansys: {
		PackageType:    Unsupported,
		PackageSubType: "Ansys",
		Icon:           iconInfo.Code,
	},
	fileInfo.BAM: {
		PackageType:    Unsupported,
		PackageSubType: "Genomics",
		Icon:           iconInfo.Genomics,
	},
	fileInfo.CRAM: {
		PackageType:    Unsupported,
		PackageSubType: "Genomics",
		Icon:           iconInfo.Genomics,
	},
	fileInfo.BIODAC: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.BioPAC: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.COMSOL: {
		PackageType:    Unsupported,
		PackageSubType: "Model",
		Icon:           iconInfo.Model,
	},
	fileInfo.CPlusPlus: {
		PackageType:    Unsupported,
		PackageSubType: "C++",
		Icon:           iconInfo.Code,
	},
	fileInfo.CSharp: {
		PackageType:    Unsupported,
		PackageSubType: "C#",
		Icon:           iconInfo.Code,
	},
	fileInfo.Data: {
		PackageType:    Unsupported,
		PackageSubType: "generic",
		Icon:           iconInfo.GenericData,
	},
	fileInfo.Docker: {
		PackageType:    Unsupported,
		PackageSubType: "Docker",
		Icon:           iconInfo.Docker,
	},
	fileInfo.EPS: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Image,
	},
	fileInfo.FCS: {
		PackageType:    Unsupported,
		PackageSubType: "Flow",
		Icon:           iconInfo.Flow,
	},
	fileInfo.FASTA: {
		PackageType:    Unsupported,
		PackageSubType: "Tabular",
		Icon:           iconInfo.Genomics,
	},
	fileInfo.FASTQ: {
		PackageType:    Unsupported,
		PackageSubType: "Tabular",
		Icon:           iconInfo.Genomics,
	},
	fileInfo.FreesurferSurface: {
		PackageType:    Unsupported,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileInfo.HDF: {
		PackageType:    Unsupported,
		PackageSubType: "Data Container",
		Icon:           iconInfo.HDF,
	},
	fileInfo.HTML: {
		PackageType:    Unsupported,
		PackageSubType: "HTML",
		Icon:           iconInfo.Code,
	},
	fileInfo.Imaris: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileInfo.Intan: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.IVCurveData: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.JAVA: {
		PackageType:    Unsupported,
		PackageSubType: "JAVA",
		Icon:           iconInfo.Code,
	},
	fileInfo.Javascript: {
		PackageType:    Unsupported,
		PackageSubType: "Javascript",
		Icon:           iconInfo.Code,
	},
	fileInfo.Json: {
		PackageType:    Unsupported,
		PackageSubType: "JSON",
		Icon:           iconInfo.JSON,
	},
	fileInfo.Jupyter: {
		PackageType:    Unsupported,
		PackageSubType: "Notebook",
		Icon:           iconInfo.Notebook,
	},
	fileInfo.LabChart: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.Leica: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileInfo.MATLAB: {
		PackageType:    HDF5,
		PackageSubType: "Data Container",
		Icon:           iconInfo.Matlab,
	},
	fileInfo.MatlabFigure: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Matlab,
	},
	fileInfo.Markdown: {
		PackageType:    Unsupported,
		PackageSubType: "Markdown",
		Icon:           iconInfo.Code,
	},
	fileInfo.Minitab: {
		PackageType:    Unsupported,
		PackageSubType: "Generic",
		Icon:           iconInfo.GenericData,
	},
	fileInfo.Neuralynx: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.NeuroDataWithoutBorders: {
		PackageType:    HDF5,
		PackageSubType: "Data Container",
		Icon:           iconInfo.NWB,
	},
	fileInfo.Neuron: {
		PackageType:    Unsupported,
		PackageSubType: "Code",
		Icon:           iconInfo.Code,
	},
	fileInfo.NihonKoden: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.Nikon: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileInfo.PatchMaster: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.PClamp: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.Plexon: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileInfo.PowerPoint: {
		PackageType:    Unsupported,
		PackageSubType: "MS Powerpoint",
		Icon:           iconInfo.PowerPoint,
	},
	fileInfo.Python: {
		PackageType:    Unsupported,
		PackageSubType: "Python",
		Icon:           iconInfo.Code,
	},
	fileInfo.R: {
		PackageType:    Unsupported,
		PackageSubType: "R",
		Icon:           iconInfo.Code,
	},
	fileInfo.RData: {
		PackageType:    Unsupported,
		PackageSubType: "Data Container",
		Icon:           iconInfo.RData,
	},
	fileInfo.Shell: {
		PackageType:    Unsupported,
		PackageSubType: "Shell",
		Icon:           iconInfo.Code,
	},
	fileInfo.SolidWorks: {
		PackageType:    Unsupported,
		PackageSubType: "Model",
		Icon:           iconInfo.Model,
	},
	fileInfo.VariantData: {
		PackageType:    Unsupported,
		PackageSubType: "Tabular",
		Icon:           iconInfo.GenomicsVariant,
	},
	fileInfo.XML: {
		PackageType:    Unsupported,
		PackageSubType: "XML",
		Icon:           iconInfo.XML,
	},
	fileInfo.YAML: {
		PackageType:    Unsupported,
		PackageSubType: "YAML",
		Icon:           iconInfo.Code,
	},
	fileInfo.ZIP: {
		PackageType:    ZIP,
		PackageSubType: "ZIP",
		Icon:           iconInfo.Zip,
	},
	fileInfo.HDF5: {
		PackageType:    HDF5,
		PackageSubType: "Data Container",
		Icon:           iconInfo.HDF,
	},
	fileInfo.Unknown: {
		PackageType:    Unknown,
		PackageSubType: "Generic",
		Icon:           iconInfo.Generic,
	},
}
