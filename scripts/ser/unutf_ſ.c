#include <stdio.h>

int nprintfiles = 0;

void printfile(char *s)
{
	int c;
	FILE *f = s ? fopen(s, "r"): stdin;

	if (f == NULL){
		perror(s);
		return;
	}

	while ((c=getc(f)) != EOF) {
		if (c < 128) {
			putchar(c);
		} else if ((c & 0xc0) == 0xc0) { /* 110xxxxx 10xxxxxx */
			int d = getc(f);
			if (c == 0xc5 && d == 0xbf) { /* s/ſ/s/ */
				putchar('s');
			}
		} else if ((c & 0xe0) == 0xe0) { /* 1110xxxx 10xxxxxx 10xxxxxx */
			getc(f);
			getc(f);
		} else if ((c & 0xf0) == 0xf0) { /* 11110xxx 10xxxxxx 10xxxxxx 10xxxxxx */
			getc(f);
			getc(f);
			getc(f);
		}
	}
	fclose(f);
}

int main(argc, argv)
	char *argv[];
{
	while (argc > 1) {
		--argc; argv++;
		if (argv[0][0] == '-'){
			switch(argv[0][1]){
			}
		} else {
			printfile(argv[0]);
			nprintfiles++;
		}
	}
	if (nprintfiles == 0)
		printfile(NULL);
}
