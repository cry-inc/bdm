function getPackageSize(manifest) {
	let size = 0;
	manifest.Files.forEach(file => {
		size += file.Object.Size;
	});
	return size;
}

function addFileNames(manifest) {
	manifest.Files.forEach(file => {
		const lastSlash = file.Path.lastIndexOf('/');
		file.Name = lastSlash !== -1 ?
			file.Path.substr(lastSlash + 1) :
			file.Path;
	});
}

function getSizeString(bytes) {
	if (bytes < 1024) {
		return bytes + " byte";
	} else if (bytes < 1000 * 1000) {
		return Math.round(bytes / 1000) + " kB";
	} else if (bytes < 1000 * 1000 * 1000) {
		return Math.round(bytes / (1000 * 1000)) + " MB";
	} else {
		return Math.round(bytes / (1000 * 1000 * 1000)) + " GB";
	}
}

export default {getPackageSize, addFileNames, getSizeString};
