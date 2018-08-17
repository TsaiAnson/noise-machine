const quilt = require('@quilt/quilt');

const deployment = quilt.createDeployment({namespace: "tsaianson-noise-machine", adminACL: ['0.0.0.0/0']});

var machine0 = new quilt.Machine({
    provider: "Amazon",
    size: "m4.large",
    preemptible: true,
    sshKeys: quilt.githubKeys('TsaiAnson'),
});

var machine1 = new quilt.Machine({
    provider: "Amazon",
    size: "m4.large",
    preemptible: true,
    sshKeys: quilt.githubKeys('TsaiAnson'),
});

// Set up noise server and machine apps
var noiseserverapp = new quilt.Container("noiseserver", "tsaianson/noise-server");

var noisemachineapp = new quilt.Container("noisemachine", "tsaianson/noise-machine").withEnv({"CPU": "0", "NET": "2", "DISK": "0", "MEM": "0"});

// Connecting server and app together
noiseserverapp.allowFrom(noisemachineapp, 80);
noisemachineapp.allowFrom(noiseserverapp, 80);

deployment.deploy(machine0.asMaster());
deployment.deploy(machine1.asWorker());
deployment.deploy(noiseserverapp);
deployment.deploy(noisemachineapp);