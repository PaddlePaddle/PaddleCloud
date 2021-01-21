# Notes

- The Dockerfile uses a develop version of paddle. If you build the image yourself, the image may not work if the version changed and is not backward compatible. Therefore the files in folder `images` are for reference purpose. Use the prebuilt image `ruminateer/paddleworkload:develop` as in `exampletj.yaml`.

- As pod discovery isn't implemented yet, the image will assume all pods in the default namespaces belong to the training job. You will need to clear the default namespace before you run the example.

- Some specs in the TrainingJob CRD may not be used by the paddle fleet code inside the image.
