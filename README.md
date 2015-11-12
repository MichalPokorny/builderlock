builderlock
======================

it locks the builder

runs on RedHat OpenShift

run `./run_local.sh` to run locally

Parts of RedHat's README.md from the cartridge template
------

Any log output will be generated to <code>$OPENSHIFT_GO_LOG_DIR</code> on your OpenShift gear


Build
-----

When you push code to your repo, a Git postreceive hook runs and invokes a compile script.  This attempts to download the Go compiler environment for you into $OPENSHIFT_GO_DIR/cache.  Once the environment is setup, the cart runs

    go get -tags openshift ./...

on a working copy of your source. 
The main file that you run will have access to two environment variables, $HOST and $PORT, which contain the internal address you must listen on to receive HTTP requests to your application.

