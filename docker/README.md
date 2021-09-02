# HRI Management API Docker Image

## Building the Image Locally
Build the code targeting linux: 
```shell script
cd src
env GOOS=linux GOACH=amd64 go build
```

Go back to the base project directory and create the image:
```shell script
cd ..
docker build ./ -f docker/Dockerfile
```
If you have an error that's asking you "Is the docker daemon running?", and it is, then why don't you go catch it?  If it's not, start up Docker by simply opening the Docker application on your computer.

## Testing
You can test locally by logging into the IBM Cloud container registry and running the docker container.
If you are making changes to your local hri-mgmt-api source code (excluding files not pushed to your branch) as you test, you will need to rebuild the image every time you want to run the container with those changes.
If you come across any trouble with this, please see the Troubleshooting section.

### Docker Run
When running the container, you will need to pass in a [config file](https://github.com/Alvearie/hri-mgmt-api/blob/main/config.yml).
If you would like to use TLS, you will need a cert and key as well.
Mount the file or directory of files you need into the container with the volume flag, `-v`.
At the end of the run command is where you add your arguments/flags.
You should just need one to specify where your config file is.  (You may include more depending on what you're testing.)

docker run command template with a mounted config file & no TLS:
```shell script
docker run --rm -p 1323:1323 -v ~/[path to your config file]:/mgmt-api-release/config/config.yml <image> -config-path ./config/config.yml
```

docker run command template with multiple mounted files:
```shell script
docker run --rm -p 1323:1323 -v ~/[path to a directory of the files to mount]:/mgmt-api-release/mounted <image> -config-path ./mounted/config.yml
```
This will put the files in your specified directory into a directory called "mounted" in the container, and the argument at the end states that the config file name is `config.yml`.
Feel free to change the name of the directory.

\
If docker run was successful, you should be able to curl your local endpoint the same way you would when running the executable locally.
As a reminder, if you are using TLS you will need to add `https://` to the beginning of your URL, and you may want to use the `-k` flag if you are using a self-signed certificate.

curling the endpoint _without_ TLS:
```shell script
curl localhost:1323/hri/healthcheck
```

curling the endpoint _with_ TLS:
```shell script
curl -k https://localhost:1323/hri/healthcheck
```

## Troubleshooting
One useful tool is running interactively with a bash prompt by adding `-it --entrypoint bash`.
This will run everything in your `docker run` command aside from the executable, and then start a bash prompt in the container directory.
This can be helpful for making sure your mounted files are inserted in the correct place.

Interactive docker run command template with no mounted files:
```shell script
docker run --rm -it --entrypoint bash <image>
```
You can add your mounted files using the same -v flag from before:
```shell script
docker run --rm -v ~/mgmt-api/myConfig.yml:/mgmt-api-release/config/myConfig.yml -it --entrypoint bash <image>
```
