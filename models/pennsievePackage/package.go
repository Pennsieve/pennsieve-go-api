package pennsievePackage

import (
	"github.com/pennsieve/pennsieve-go-api/models/icon"
	file "github.com/pennsieve/pennsieve-go-api/models/pennsieveFile"
)

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
var FileTypeDict = map[file.Type]FileTypeInfo{
	file.MEF: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.EDF: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.TDMS: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.OpenEphys: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.Persyst: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.NeuroExplorer: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.MobergSeries: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.BFTS: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.Nicolet: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.MEF3: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.Feather: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.NEV: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.Spike2: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.MINC: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           icon.ClinicalImageBrain,
	},
	file.DICOM: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           icon.ClinicalImageBrain,
	},
	file.NIFTI: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           icon.ClinicalImageBrain,
	},
	file.ROI: {
		PackageType:    Unsupported,
		PackageSubType: "Morphology",
		Icon:           icon.ClinicalImageBrain,
	},
	file.SWC: {
		PackageType:    Unsupported,
		PackageSubType: "Morphology",
		Icon:           icon.ClinicalImageBrain,
	},
	file.ANALYZE: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           icon.ClinicalImageBrain,
	},
	file.MGH: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           icon.ClinicalImageBrain,
	},
	file.JPEG: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           icon.Image,
	},
	file.PNG: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           icon.Image,
	},
	file.TIFF: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	file.OMETIFF: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	file.BRUKERTIFF: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	file.CZI: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	file.JPEG2000: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	file.LSM: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	file.NDPI: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	file.OIB: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	file.OIF: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	file.GIF: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           icon.Image,
	},
	file.WEBM: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           icon.Video,
	},
	file.MOV: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           icon.Video,
	},
	file.AVI: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           icon.Video,
	},
	file.MP4: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           icon.Video,
	},
	file.CSV: {
		PackageType:    CSV,
		PackageSubType: "Tabular",
		Icon:           icon.Tabular,
	},
	file.TSV: {
		PackageType:    CSV,
		PackageSubType: "Tabular",
		Icon:           icon.Tabular,
	},
	file.MSExcel: {
		PackageType:    Unsupported,
		PackageSubType: "MS Excel",
		Icon:           icon.Excel,
	},
	file.Aperio: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	file.MSWord: {
		PackageType:    MSWord,
		PackageSubType: "MS Word",
		Icon:           icon.Word,
	},
	file.PDF: {
		PackageType:    PDF,
		PackageSubType: "PDF",
		Icon:           icon.PDF,
	},
	file.Text: {
		PackageType:    Text,
		PackageSubType: "Text",
		Icon:           icon.Text,
	},
	file.BFANNOT: {
		PackageType:    Unknown,
		PackageSubType: "Text",
		Icon:           icon.Generic,
	},
	file.AdobeIllustrator: {
		PackageType:    Unsupported,
		PackageSubType: "Illustrator",
		Icon:           icon.AdobeIllustrator,
	},
	file.AFNI: {
		PackageType:    Unsupported,
		PackageSubType: "3D Image",
		Icon:           icon.ClinicalImageBrain,
	},
	file.AFNIBRIK: {
		PackageType:    Unsupported,
		PackageSubType: "3D Image",
		Icon:           icon.ClinicalImageBrain,
	},
	file.Ansys: {
		PackageType:    Unsupported,
		PackageSubType: "Ansys",
		Icon:           icon.Code,
	},
	file.BAM: {
		PackageType:    Unsupported,
		PackageSubType: "Genomics",
		Icon:           icon.Genomics,
	},
	file.CRAM: {
		PackageType:    Unsupported,
		PackageSubType: "Genomics",
		Icon:           icon.Genomics,
	},
	file.BIODAC: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.BioPAC: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.COMSOL: {
		PackageType:    Unsupported,
		PackageSubType: "Model",
		Icon:           icon.Model,
	},
	file.CPlusPlus: {
		PackageType:    Unsupported,
		PackageSubType: "C++",
		Icon:           icon.Code,
	},
	file.CSharp: {
		PackageType:    Unsupported,
		PackageSubType: "C#",
		Icon:           icon.Code,
	},
	file.Data: {
		PackageType:    Unsupported,
		PackageSubType: "generic",
		Icon:           icon.GenericData,
	},
	file.Docker: {
		PackageType:    Unsupported,
		PackageSubType: "Docker",
		Icon:           icon.Docker,
	},
	file.EPS: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Image,
	},
	file.FCS: {
		PackageType:    Unsupported,
		PackageSubType: "Flow",
		Icon:           icon.Flow,
	},
	file.FASTA: {
		PackageType:    Unsupported,
		PackageSubType: "Tabular",
		Icon:           icon.Genomics,
	},
	file.FASTQ: {
		PackageType:    Unsupported,
		PackageSubType: "Tabular",
		Icon:           icon.Genomics,
	},
	file.FreesurferSurface: {
		PackageType:    Unsupported,
		PackageSubType: "3D Image",
		Icon:           icon.ClinicalImageBrain,
	},
	file.HDF: {
		PackageType:    Unsupported,
		PackageSubType: "Data Container",
		Icon:           icon.HDF,
	},
	file.HTML: {
		PackageType:    Unsupported,
		PackageSubType: "HTML",
		Icon:           icon.Code,
	},
	file.Imaris: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	file.Intan: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.IVCurveData: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.JAVA: {
		PackageType:    Unsupported,
		PackageSubType: "JAVA",
		Icon:           icon.Code,
	},
	file.Javascript: {
		PackageType:    Unsupported,
		PackageSubType: "Javascript",
		Icon:           icon.Code,
	},
	file.Json: {
		PackageType:    Unsupported,
		PackageSubType: "JSON",
		Icon:           icon.JSON,
	},
	file.Jupyter: {
		PackageType:    Unsupported,
		PackageSubType: "Notebook",
		Icon:           icon.Notebook,
	},
	file.LabChart: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.Leica: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	file.MATLAB: {
		PackageType:    HDF5,
		PackageSubType: "Data Container",
		Icon:           icon.Matlab,
	},
	file.MatlabFigure: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Matlab,
	},
	file.Markdown: {
		PackageType:    Unsupported,
		PackageSubType: "Markdown",
		Icon:           icon.Code,
	},
	file.Minitab: {
		PackageType:    Unsupported,
		PackageSubType: "Generic",
		Icon:           icon.GenericData,
	},
	file.Neuralynx: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.NeuroDataWithoutBorders: {
		PackageType:    HDF5,
		PackageSubType: "Data Container",
		Icon:           icon.NWB,
	},
	file.Neuron: {
		PackageType:    Unsupported,
		PackageSubType: "Code",
		Icon:           icon.Code,
	},
	file.NihonKoden: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.Nikon: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           icon.Microscope,
	},
	file.PatchMaster: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.PClamp: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.Plexon: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           icon.Timeseries,
	},
	file.PowerPoint: {
		PackageType:    Unsupported,
		PackageSubType: "MS Powerpoint",
		Icon:           icon.PowerPoint,
	},
	file.Python: {
		PackageType:    Unsupported,
		PackageSubType: "Python",
		Icon:           icon.Code,
	},
	file.R: {
		PackageType:    Unsupported,
		PackageSubType: "R",
		Icon:           icon.Code,
	},
	file.RData: {
		PackageType:    Unsupported,
		PackageSubType: "Data Container",
		Icon:           icon.RData,
	},
	file.Shell: {
		PackageType:    Unsupported,
		PackageSubType: "Shell",
		Icon:           icon.Code,
	},
	file.SolidWorks: {
		PackageType:    Unsupported,
		PackageSubType: "Model",
		Icon:           icon.Model,
	},
	file.VariantData: {
		PackageType:    Unsupported,
		PackageSubType: "Tabular",
		Icon:           icon.GenomicsVariant,
	},
	file.XML: {
		PackageType:    Unsupported,
		PackageSubType: "XML",
		Icon:           icon.XML,
	},
	file.YAML: {
		PackageType:    Unsupported,
		PackageSubType: "YAML",
		Icon:           icon.Code,
	},
	file.ZIP: {
		PackageType:    ZIP,
		PackageSubType: "ZIP",
		Icon:           icon.Zip,
	},
	file.HDF5: {
		PackageType:    HDF5,
		PackageSubType: "Data Container",
		Icon:           icon.HDF,
	},
	file.Unknown: {
		PackageType:    Unknown,
		PackageSubType: "Generic",
		Icon:           icon.Generic,
	},
}
