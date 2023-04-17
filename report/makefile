REPORT_NAME:=STU_2022-23_Individual_Project_40_LeoSpratt

help:
	@echo "Usage:"
	@echo
	@echo "  build    make a pdf and include refs"
	@echo "  biber    include refs from generated tex"
	@echo "  pdf      build a pdf"
	@echo "  clean    remove all generated files (including pdf)"
build: pdf biber glossaries pdf pdf
biber:
	@biber ${REPORT_NAME}
glossaries:
	@makeglossaries ${REPORT_NAME}
pdf:
	@lualatex --jobname=${REPORT_NAME} report.tex
clean:
	@rm ${REPORT_NAME}.*
