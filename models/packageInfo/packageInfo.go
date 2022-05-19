package packageInfo

import (
	"github.com/pennsieve/pennsieve-go-api/models/fileInfo"
	"github.com/pennsieve/pennsieve-go-api/models/icon"
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

type FileTypeInfo struct {
	PackageType    Type
	PackageSubType string
	Icon           icon.Icon
}

// fileTypeDict maps filetypes to PackageTypes.
var FileTypeDict = map[fileInfo.Type]FileTypeInfo{
	fileInfo.MEF: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.EDF: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.TDMS: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.OpenEphys: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.Persyst: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.NeuroExplorer: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.MobergSeries: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.BFTS: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.Nicolet: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.MEF3: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.Feather: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.NEV: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.Spike2: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.MINC: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           icon.ClinicalImageBrain,
	},
	fileInfo.DICOM: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           icon.ClinicalImageBrain,
	},
	fileInfo.NIFTI: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           icon.ClinicalImageBrain,
	},
	fileInfo.ROI: {
		PackageType:    Unsupported,
		PackageSubType: "Morphology",
		Icon:           icon.ClinicalImageBrain,
	},
	fileInfo.SWC: {
		PackageType:    Unsupported,
		PackageSubType: "Morphology",
		Icon:           icon.ClinicalImageBrain,
	},
	fileInfo.ANALYZE: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           icon.ClinicalImageBrain,
	},
	fileInfo.MGH: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           icon.ClinicalImageBrain,
	},
	fileInfo.JPEG: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           icon.Image,
	},
	fileInfo.PNG: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           icon.Image,
	},
	fileInfo.TIFF: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	fileInfo.OMETIFF: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	fileInfo.BRUKERTIFF: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	fileInfo.CZI: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	fileInfo.JPEG2000: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	fileInfo.LSM: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	fileInfo.NDPI: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	fileInfo.OIB: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	fileInfo.OIF: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	fileInfo.GIF: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           icon.Image,
	},
	fileInfo.WEBM: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           icon.Video,
	},
	fileInfo.MOV: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           icon.Video,
	},
	fileInfo.AVI: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           icon.Video,
	},
	fileInfo.MP4: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           icon.Video,
	},
	fileInfo.CSV: {
		PackageType:    CSV,
		PackageSubType: "Tabular",
		Icon:           icon.Tabular,
	},
	fileInfo.TSV: {
		PackageType:    CSV,
		PackageSubType: "Tabular",
		Icon:           icon.Tabular,
	},
	fileInfo.MSExcel: {
		PackageType:    Unsupported,
		PackageSubType: "MS Excel",
		Icon:           icon.Excel,
	},
	fileInfo.Aperio: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	fileInfo.MSWord: {
		PackageType:    MSWord,
		PackageSubType: "MS Word",
		Icon:           icon.Word,
	},
	fileInfo.PDF: {
		PackageType:    PDF,
		PackageSubType: "PDF",
		Icon:           icon.PDF,
	},
	fileInfo.Text: {
		PackageType:    Text,
		PackageSubType: "Text",
		Icon:           icon.Text,
	},
	fileInfo.BFANNOT: {
		PackageType:    Unknown,
		PackageSubType: "Text",
		Icon:           icon.Generic,
	},
	fileInfo.AdobeIllustrator: {
		PackageType:    Unsupported,
		PackageSubType: "Illustrator",
		Icon:           icon.AdobeIllustrator,
	},
	fileInfo.AFNI: {
		PackageType:    Unsupported,
		PackageSubType: "3D Image",
		Icon:           icon.ClinicalImageBrain,
	},
	fileInfo.AFNIBRIK: {
		PackageType:    Unsupported,
		PackageSubType: "3D Image",
		Icon:           icon.ClinicalImageBrain,
	},
	fileInfo.Ansys: {
		PackageType:    Unsupported,
		PackageSubType: "Ansys",
		Icon:           icon.Code,
	},
	fileInfo.BAM: {
		PackageType:    Unsupported,
		PackageSubType: "Genomics",
		Icon:           icon.Genomics,
	},
	fileInfo.CRAM: {
		PackageType:    Unsupported,
		PackageSubType: "Genomics",
		Icon:           icon.Genomics,
	},
	fileInfo.BIODAC: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.BioPAC: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.COMSOL: {
		PackageType:    Unsupported,
		PackageSubType: "Model",
		Icon:           icon.Model,
	},
	fileInfo.CPlusPlus: {
		PackageType:    Unsupported,
		PackageSubType: "C++",
		Icon:           icon.Code,
	},
	fileInfo.CSharp: {
		PackageType:    Unsupported,
		PackageSubType: "C#",
		Icon:           icon.Code,
	},
	fileInfo.Data: {
		PackageType:    Unsupported,
		PackageSubType: "generic",
		Icon:           icon.GenericData,
	},
	fileInfo.Docker: {
		PackageType:    Unsupported,
		PackageSubType: "Docker",
		Icon:           icon.Docker,
	},
	fileInfo.EPS: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Image,
	},
	fileInfo.FCS: {
		PackageType:    Unsupported,
		PackageSubType: "Flow",
		Icon:           icon.Flow,
	},
	fileInfo.FASTA: {
		PackageType:    Unsupported,
		PackageSubType: "Tabular",
		Icon:           icon.Genomics,
	},
	fileInfo.FASTQ: {
		PackageType:    Unsupported,
		PackageSubType: "Tabular",
		Icon:           icon.Genomics,
	},
	fileInfo.FreesurferSurface: {
		PackageType:    Unsupported,
		PackageSubType: "3D Image",
		Icon:           icon.ClinicalImageBrain,
	},
	fileInfo.HDF: {
		PackageType:    Unsupported,
		PackageSubType: "Data Container",
		Icon:           icon.HDF,
	},
	fileInfo.HTML: {
		PackageType:    Unsupported,
		PackageSubType: "HTML",
		Icon:           icon.Code,
	},
	fileInfo.Imaris: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	fileInfo.Intan: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.IVCurveData: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.JAVA: {
		PackageType:    Unsupported,
		PackageSubType: "JAVA",
		Icon:           icon.Code,
	},
	fileInfo.Javascript: {
		PackageType:    Unsupported,
		PackageSubType: "Javascript",
		Icon:           icon.Code,
	},
	fileInfo.Json: {
		PackageType:    Unsupported,
		PackageSubType: "JSON",
		Icon:           icon.JSON,
	},
	fileInfo.Jupyter: {
		PackageType:    Unsupported,
		PackageSubType: "Notebook",
		Icon:           icon.Notebook,
	},
	fileInfo.LabChart: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.Leica: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	fileInfo.MATLAB: {
		PackageType:    HDF5,
		PackageSubType: "Data Container",
		Icon:           icon.Matlab,
	},
	fileInfo.MatlabFigure: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Matlab,
	},
	fileInfo.Markdown: {
		PackageType:    Unsupported,
		PackageSubType: "Markdown",
		Icon:           icon.Code,
	},
	fileInfo.Minitab: {
		PackageType:    Unsupported,
		PackageSubType: "Generic",
		Icon:           icon.GenericData,
	},
	fileInfo.Neuralynx: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.NeuroDataWithoutBorders: {
		PackageType:    HDF5,
		PackageSubType: "Data Container",
		Icon:           icon.NWB,
	},
	fileInfo.Neuron: {
		PackageType:    Unsupported,
		PackageSubType: "Code",
		Icon:           icon.Code,
	},
	fileInfo.NihonKoden: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.Nikon: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	fileInfo.PatchMaster: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.PClamp: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.Plexon: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	fileInfo.PowerPoint: {
		PackageType:    Unsupported,
		PackageSubType: "MS Powerpoint",
		Icon:           icon.PowerPoint,
	},
	fileInfo.Python: {
		PackageType:    Unsupported,
		PackageSubType: "Python",
		Icon:           icon.Code,
	},
	fileInfo.R: {
		PackageType:    Unsupported,
		PackageSubType: "R",
		Icon:           icon.Code,
	},
	fileInfo.RData: {
		PackageType:    Unsupported,
		PackageSubType: "Data Container",
		Icon:           icon.RData,
	},
	fileInfo.Shell: {
		PackageType:    Unsupported,
		PackageSubType: "Shell",
		Icon:           icon.Code,
	},
	fileInfo.SolidWorks: {
		PackageType:    Unsupported,
		PackageSubType: "Model",
		Icon:           icon.Model,
	},
	fileInfo.VariantData: {
		PackageType:    Unsupported,
		PackageSubType: "Tabular",
		Icon:           icon.GenomicsVariant,
	},
	fileInfo.XML: {
		PackageType:    Unsupported,
		PackageSubType: "XML",
		Icon:           icon.XML,
	},
	fileInfo.YAML: {
		PackageType:    Unsupported,
		PackageSubType: "YAML",
		Icon:           icon.Code,
	},
	fileInfo.ZIP: {
		PackageType:    ZIP,
		PackageSubType: "ZIP",
		Icon:           icon.Zip,
	},
	fileInfo.HDF5: {
		PackageType:    HDF5,
		PackageSubType: "Data Container",
		Icon:           icon.HDF,
	},
	fileInfo.Unknown: {
		PackageType:    Unknown,
		PackageSubType: "Generic",
		Icon:           icon.Generic,
	},
}
