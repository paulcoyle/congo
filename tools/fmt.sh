TEMPFILE="__gofmt.temp"
for f in `find . -name "*.go"`; do
  echo "Formatting $f..."
  gofmt -tabs=false -tabwidth=2 $f > $TEMPFILE
  mv $TEMPFILE $f
done
echo "Done!"
