export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:./lib:./
./AIservice -m=0 -c=aiges.toml -s=mocksvc -u=http://10.1.87.70:6868 -p=guiderAllService -g=gas
