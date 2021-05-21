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

export {getPackageSize, addFileNames};
