.DEFAULT: protos

protos: dependency generate

dependency:
	sh scripts/download_dep.sh

generate:
	sh scripts/generate_proto.sh

