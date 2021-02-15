#include <stdio.h>

#include "../clib/bdm.h"

int main() {
	char* packageName = "foo";
	char* outputFolder = ".\\out";
	char* serverURL = "http://127.0.0.1:2323";
	int packageVersion = 2;
	int clean = 1;

	int result = bdmDownloadPackage(packageName, packageVersion, outputFolder, serverURL, clean);
	
	printf("bdmDownloadPackage returned %d\n", result);
	return 0;
}
