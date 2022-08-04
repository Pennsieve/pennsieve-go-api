package handler

import (
	"github.com/pennsieve/pennsieve-go-api/models/fileInfo/fileType"
	"github.com/pennsieve/pennsieve-go-api/models/manifest/manifestFile"
	"log"
	"regexp"
	"strings"
)

func (s ManifestSession) PackageTypeResolver(items []manifestFile.FileDTO) []manifestFile.FileDTO {

	for i, f := range items {

		// Determine Type based on extension, or
		// return type that is already defined in FileDTO
		var fileName, fileExtension string
		var fType fileType.Type
		if len(items[i].FileType) == 0 {
			// 1. Find FileType

			log.Println("FILETYPE IS NOT SET")

			r := regexp.MustCompile(`(?P<FileName>[^\.]*)?\.?(?P<Extension>.*)`)
			pathParts := r.FindStringSubmatch(f.TargetName)
			if pathParts == nil {
				log.Println("Unable to parse filename:", f.TargetName)
				continue
			}

			fileName = pathParts[r.SubexpIndex("FileName")]
			fileExtension = pathParts[r.SubexpIndex("Extension")]
			fType = fileType.ExtensionToTypeDict[fileExtension]

			// Set the type if not previously set.
			items[i].FileType = fType.String()
		} else {

			log.Println("FILETYPE IS SET")
			fType = fileType.Dict[items[i].FileType]
		}

		log.Println("File type: ", items[i].FileType)

		// 2. Implement Merge Strategy
		switch fType {
		case fileType.Persyst:
			persystMerger(fileName, &items[i], items)
		default:
			continue

		}
	}
	return items
}

func persystMerger(fileName string, layFile *manifestFile.FileDTO, items []manifestFile.FileDTO) {

	// Iterate over files and if file exists in same folder with same name and ".dat" extension, than merge the two.
	// Then set MergePackageID for both lay and dat file to the uploadID of the lay file.
	// This ensures that when we create the package in upload_handler that we set the name of the package to be
	// the filename without extension.
	for i, f := range items {
		if layFile.TargetPath == f.TargetPath && layFile.TargetName != f.TargetName {
			if strings.HasPrefix(f.TargetName, fileName) && strings.HasSuffix(f.TargetName, ".dat") {
				items[i].MergePackageId = layFile.UploadID
				layFile.MergePackageId = layFile.UploadID
				items[i].FileType = fileType.Persyst.String()
				log.Println("Found match in: ", f.TargetName)
			}
			break
		}
	}
}
