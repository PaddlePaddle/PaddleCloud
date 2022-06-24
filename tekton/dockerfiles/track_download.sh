OUT_FILE='log'

OCR_DOWNLOAD=$(curl -s https://hub.docker.com/v2/repositories/paddlecloud/paddleocr/ | jq -r ".pull_count")
NLP_DOWNLOAD=$(curl -s https://hub.docker.com/v2/repositories/paddlecloud/paddlenlp/ | jq -r ".pull_count")
DETECTION_DOWNLOAD=$(curl -s https://hub.docker.com/v2/repositories/paddlecloud/paddledetection/ | jq -r ".pull_count")
SEG_DOWNLOAD=$(curl -s https://hub.docker.com/v2/repositories/paddlecloud/paddleseg/ | jq -r ".pull_count")
CLAS_DOWNLOAD=$(curl -s https://hub.docker.com/v2/repositories/paddlecloud/paddleclas/ | jq -r ".pull_count")
REC_DOWNLOAD=$(curl -s https://hub.docker.com/v2/repositories/paddlecloud/paddlerec/ | jq -r ".pull_count")
SPEECH_DOWNLOAD=$(curl -s https://hub.docker.com/v2/repositories/paddlecloud/paddlespeech/ | jq -r ".pull_count")
COLLECTION_DOWNLOAD=$(curl -s https://hub.docker.com/v2/repositories/paddlecloud/paddle-toolkit-collection/ | jq -r ".pull_count")

if [ ! -f "${OUT_FILE}" ]; then
  echo 'hello'
  TITLE="Time \t Collection \t OCR \t NLP \t Detection \t Seg \t Clas \t Speech \t Rec"
  touch ${OUT_FILE}
  echo "${TITLE}" >> ${OUT_FILE}
fi
TIME=$(date "+%m-%d")

LINE="${TIME} \t ${COLLECTION_DOWNLOAD} \t ${OCR_DOWNLOAD} \t ${NLP_DOWNLOAD} \t ${DETECTION_DOWNLOAD} \t ${SEG_DOWNLOAD} \t ${CLAS_DOWNLOAD} \t ${SPEECH_DOWNLOAD} \t ${REC_DOWNLOAD}"
echo "${LINE}" >> ${OUT_FILE}
