#!/bin/bash

set -e

OEMDIR=/oem
FEATURESDIR=/system/features

usage()
{
    echo "Usage: cos-feature [list|enable|disable] <feature>"
    echo ""
    echo "Example: cos-feature enable k3s"
    echo ""
    echo "To list available features, run: cos-feature list"
    echo "To enable, run: cos-feature enable <feature>"
    echo "To disable, run: cos-feature disable <feature>"
    echo ""
    exit 1
}

list() {
  echo ""
  echo "===================="
  echo "cOS features list"
  echo ""
  echo "To enable, run: cos-feature enable <feature>"
  echo "To disable, run: cos-feature disable <feature>"
  echo "===================="
  echo ""
  for i in $FEATURESDIR/*.yaml; do
    f=$(basename $i .yaml)
    if [ -L "$OEMDIR/features/${f}.yaml" ]; then
      enabled="(enabled)"
    fi
    echo "- $f $enabled"
  done

}

enable() {
  if [ ! -e "$FEATURESDIR/$1.yaml" ]; then
    echo "Feature not present"
    exit 1
  fi
  if [ ! -d "$OEMDIR/features" ]; then
    mkdir $OEMDIR/features
  fi
  ln -s $FEATURESDIR/$1.yaml $OEMDIR/features/$1.yaml
  echo "$1 enabled"
}

disable() {
  if [ -L "$OEMDIR/features/$1.yaml" ]; then
    rm $OEMDIR/features/$1.yaml
    echo "Feature $1 disabled"
  else
    echo "Feature $1 not enabled"
  fi
}

while [ "$#" -gt 0 ]; do
    case $1 in
        list)
            shift 1
            list
            ;;
        enable)
            shift 1
            enable $1
            ;;
        disable)
            shift 1
            disable $1
            ;;
        -h)
            usage
            ;;
        --help)
            usage
            ;;
        *)
            if [ "$#" -gt 2 ]; then
                usage
            fi
            INTERACTIVE=true
            break
            ;;
    esac
done
