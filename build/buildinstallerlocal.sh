#!/bin/bash

# this file is to automate the steps that will be performed by buildsever
# on the local machine to make an installer

# to reuse the script, modify , version, name file and path to copy etc

# create a workspace directory under ~/.tmp using version and name
# Start from the build directory.. traverse to main binary directory and build
# make a zip file with contents of bin
# copy all files 
# call the makebin-auto.sh file 

VERSION=0.1
COMPONENTNAME=workloadservice
COMPONENT=$COMPONENTNAME-$VERSION

WORKSPACEDIR=~/.tmp/workspace/$COMPONENTNAME-$VERSION
BUILDOUTDIR=$WORKSPACEDIR/buildout
TARGETDIR=`pwd`/target

# make a clean workspace 

rm -rf $WORKSPACEDIR
rm -rf $TARGETDIR
mkdir -p $WORKSPACEDIR
mkdir -p $TARGETDIR
ls -l $WORKSPACEDIR


# move to the binary directory temporarily to build the binary
cd ../
go build -o $BUILDOUTDIR/bin/workloadservice.bin

# move back to working directory
cd -
# do any other builds and store in the bin directory
cd $BUILDOUTDIR
zip -r $WORKSPACEDIR/$COMPONENT.zip .
unzip -l $WORKSPACEDIR/$COMPONENT.zip
#tar -cvzf $WORKSPACEDIR/$COMPONENT.tar.gz -C $BUILDOUTDIR .
#tar -tvf $WORKSPACEDIR/$COMPONENT.tar.gz
cd -
#delete
rm -rf $BUILDOUTDIR

cp ../files/* $WORKSPACEDIR
cp ../etc/* $WORKSPACEDIR

. ../build/makebin-auto.sh $WORKSPACEDIR
cp $WORKSPACEDIR*.bin $TARGETDIR

rm -rf $WORKSPACEDIR