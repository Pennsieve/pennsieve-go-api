package file

// FileType is an enum indicating the type of the File
type FileType int64

const (
	PDF FileType = iota
	MEF
	EDF
	TDMS
	OpenEphys
	Persyst
	DICOM
	NIFTI
	PNG
	CZI
	Aperio
	Json
	CSV
	TSV
	Text
	XML
	HTML
	MSExcel
	MSWord
	MP4
	WEBM
	OGG
	MOV
	JPEG
	JPEG2000
	LSM
	NDPI
	OIB
	OIF
	ROI
	SWC
	CRAM
	MGH
	AVI
	MATLAB
	HDF5
	TIFF
	OMETIFF
	BRUKERTIFF
	GIF
	ANALYZE
	NeuroExplorer
	MINC
	MobergSeries
	GenericData
	BFANNOT
	BFTS
	Nicolet
	MEF3
	Feather
	NEV
	Spike2
	AdobeIllustrator
	AFNI
	AFNIBRIK
	Ansys
	BAM
	BIODAC
	BioPAC
	COMSOL
	CPlusPlus
	CSharp
	Data
	Docker
	EPS
	FCS
	FASTA
	FASTQ
	FreesurferSurface
	HDF
	Imaris
	Intan
	IVCurveData
	JAVA
	Javascript
	Jupyter
	LabChart
	Leica
	MatlabFigure
	Markdown
	Minitab
	Neuralynx
	NeuroDataWithoutBorders
	Neuron
	NihonKoden
	Nikon
	PatchMaster
	PClamp
	Plexon
	PowerPoint
	Python
	R
	RData
	Shell
	SolidWorks
	VariantData
	YAML
	ZIP
)

//func (s FileType) String() string {
//	switch s {
//	case models.Unavailable:
//		return "UNAVAILABLE"
//	case models.Uploaded:
//		return "UPLOADED"
//	case models.Deleting:
//		return "DELETING"
//	case models.Infected:
//		return "INFECTED"
//	case models.UploadFailed:
//		return "UPLOAD_FAILED"
//	case models.Processing:
//		return "PROCESSING"
//	case models.Ready:
//		return "READY"
//	case models.ProcessingFailed:
//		return "PROCESSING_FAILED"
//	}
//	return "UNKNOWN"
//}

var FileExtensionDict = map[string]FileType{
	"bfannot": BFANNOT,
	"bfts":    BFTS,
	// Image file
	"png":           PNG,
	"jpg":           JPEG,
	"jpeg":          JPEG,
	"jp2":           JPEG2000,
	"jpx":           JPEG2000,
	"lsm":           LSM,
	"ndpi":          NDPI,
	"oib":           OIB,
	"oif":           OIF,
	"ome.tiff":      OMETIFF,
	"ome.tif":       OMETIFF,
	"ome.tf2":       OMETIFF,
	"ome.tf8":       OMETIFF,
	"ome.btf":       OMETIFF,
	"brukertiff.gz": BRUKERTIFF,
	"tiff":          TIFF,
	"tif":           TIFF,
	"gif":           GIF,
	"ai":            AdobeIllustrator,
	"svg":           AdobeIllustrator,
	"nd2":           Nikon,
	"lif":           Leica,
	"ims":           Imaris,
	// Markup/Text file
	"txt":  Text,
	"text": Text,
	"rtf":  Text,
	"html": HTML,
	"htm":  HTML,
	"csv":  CSV,
	"pdf":  PDF,
	"doc":  MSWord,
	"docx": MSWord,
	"json": Json,
	"xls":  MSExcel,
	"xlsx": MSExcel,
	"xml":  XML,
	"tsv":  TSV,
	"ppt":  PowerPoint,
	"pptx": PowerPoint,
	// Matlab

	"mat": MATLAB,
	"mex": MATLAB,
	"m":   MATLAB,
	"fig": MatlabFigure,
	// ------- TimeSeries --------------
	"mef":     MEF,
	"mefd.gz": MEF3,
	"edf":     EDF,
	"tdm":     TDMS,
	"tdms":    TDMS,
	"lay":     Persyst,
	"dat":     Persyst,
	"nex":     NeuroExplorer,
	"nex5":    NeuroExplorer,
	"smr":     Spike2,
	// Nihon Kohden
	".eeg": NihonKoden,
	// Plexon
	"plx": Plexon,
	"pl2": Plexon,
	// Nicolet
	"e": Nicolet,
	//Open Ephys
	"continuous": OpenEphys,
	"spikes":     OpenEphys,
	"events":     OpenEphys,
	"openephys":  OpenEphys,
	//NEV
	"nev": NEV,
	"ns1": NEV,
	"ns2": NEV,
	"ns3": NEV,
	"ns4": NEV,
	"ns5": NEV,
	"ns6": NEV,
	"nf3": NEV,
	// Moberg
	"moberg.gz": MobergSeries,
	// Apache Feather
	"feather": Feather,
	// BIODAC
	"tab": BIODAC,
	// BioPAC
	"acq": BioPAC,
	// Intan
	"rhd": Intan,
	// IV Curve Data
	"ibw": IVCurveData,
	// LabChart
	"adicht": LabChart,
	"adidat": LabChart,
	// Neuralynx
	"ncs": Neuralynx,
	// PatchMaster
	"pgf": PatchMaster,
	"pul": PatchMaster,
	// pClamp
	"abf": PClamp,
	// ------- Imaging --------------
	"dcm":    DICOM,
	"dicom":  DICOM,
	"nii":    NIFTI,
	"nii.gz": NIFTI,
	"nifti":  NIFTI,
	"roi":    ROI,
	"swc":    SWC,
	"mgh":    MGH,
	"mgz":    MGH,
	"mgh.gz": MGH,
	"mnc":    MINC,
	"img":    ANALYZE,
	"hdr":    ANALYZE,
	"afni":   AFNI,
	"brik":   AFNIBRIK,
	"head":   AFNIBRIK,
	"lh":     FreesurferSurface,
	"rh":     FreesurferSurface,
	"curv":   FreesurferSurface,
	"eps":    EPS,
	"ps":     EPS,
	// 2d imaging
	"svs": Aperio,
	"czi": CZI,
	// ------- Video --------------
	"mov":  MOV,
	"mp4":  MP4,
	"ogg":  OGG,
	"ogv":  OGG,
	"webm": WEBM,
	"avi":  AVI,
	// 3D Model
	"mph":    COMSOL,
	"sldasm": SolidWorks,
	"slddrw": SolidWorks,
	// Aggregate
	"hdf":     HDF,
	"hdf4":    HDF,
	"hdf5":    HDF5,
	"h5":      HDF5,
	"h4":      HDF,
	"he2":     HDF,
	"he5":     HDF,
	"mpj":     Minitab,
	"mtw":     Minitab,
	"mgf":     Minitab,
	"nwb":     NeuroDataWithoutBorders,
	"rdata":   RData,
	"zip":     ZIP,
	"tar":     ZIP,
	".tar.gz": ZIP,
	// Flow
	"fcs": FCS,
	// Genomics
	"bam":    BAM,
	"bcl":    BAM,
	"bcl.gz": BAM,
	"fasta":  FASTA,
	"fastq":  FASTQ,
	"vcf":    VariantData,
	"cram":   CRAM,
	//source code
	"cs":   CSharp,
	"aedt": Ansys,
	"cpp":  CPlusPlus,
	"js":   Javascript,
	"md":   Markdown,
	"hoc":  Neuron,
	"mod":  Neuron,
	"py":   Python,
	"r":    R,
	"sh":   Shell,
	"tcsh": Shell,
	"bash": Shell,
	"zsh":  Shell,
	"yaml": YAML,
	"yml":  YAML,
	"java": JAVA,
	//Generic data
	"data": Data,
	"bin":  Data,
	"raw":  Data,
	//Other
	"Dockerfile": Docker,
	"ipynb":      Jupyter,
}
