#! /bin/sh

cmd="gen-crd-api-reference-docs";

cd $(cd "$(dirname "$0")";pwd);

if [ ! -x "$cmd" ]; then
  wget https://github.com/ahmetb/gen-crd-api-reference-docs/releases/download/v0.1.5/gen-crd-api-reference-docs_darwin_amd64.tar.gz;
  tar xzvf gen-crd-api-reference-docs_darwin_amd64.tar.gz;
  rm -rf gen-crd-api-reference-docs_darwin_amd64.tar.gz;
fi

./"$cmd" \
  --config ./example-config.json \
  --template-dir ./template \
  --api-dir ../../api \
  --out-file ../en/api_doc.md