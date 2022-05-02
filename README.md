# ns3-playground
Online Network Simulator for Internet Systems


Start the docker 

```
docker build -t ns3 .
```

```
docker run -t -d --name ns3-{random-id} ns3
docker cp ${PWD}/uploads/{random-id} ns3-{random-id}:/usr/ns-allinone-3.30.1/ns-3.30.1/scratch/file.cpp
docker exec ns3-{random-id} sh -c "cd /usr/ns-allinone-3.30.1/ns-3.30.1/ && ./waf"
docker exec ns3-{random-id} sh -c "cd /usr/ns-allinone-3.30.1/ns-3.30.1/ && ./waf --run file"
docker cp ${PWD}/uploads/{random-id} ns3-{random-id}:/usr/ns-allinone-3.30.1/ns-3.30.1/scratch/file.cpp

/usr/ns-allinone-3.30.1/ns-3.30.1
dcoker stop ns3-{random-id}
docker rm ns3-{random-id}
```

TODO:
[] Show result in the window
[] user to generate user and password and store files - ns3playground.io
[] import program template from examples folder