# Serverless-suite

A collection of serverless benchmarks and detailed elaboration on the methodology and how to run them on both the Gem5 simulator and your own machine.

## Setting up Gem5

Main site [here](https://www.gem5.org/)

```bash
git clone https://github.com/gem5/gem5.git
```

Why? Gem5 simulator allows you to perform systems research in a controllable environment, compared to real systems which are too complex and there are too many factors to control for.

Setting up Gem5 simulator for serverless workloads contains the following components

0. Disk: File system for Linux
1. Workload: linux kernel
2. Run script: What Linux should do?
    1. Start function
    2. Invoke function
3. Gem5 Config: Defines the system architecture: how many CPU cores, caches...
4. Stats: Experiment results

### Compiled Gem5 Sources

```bash
./scripts/setup_gem5.sh
```

This script installs all dependencies, pulls the gem5 repo and builds all components of gem5. This can take minutes to hours depending on machine and number of cores.

### Build Linux Kernel

The Linux kernel binary is executed on the simulated hardware system. The kernel needs certain modules for gem5 to properly work and can be found [here](https://gem5.googlesource.com/public/gem5-resources/+/refs/heads/stable/src/linux-kernel/). Alternatively, use the provided config in `/config` and the script `/scripts/setup_kernel.sh`.

```bash
# Build kernel for gem5 supporting containerized workloads
KVERSION=5.4.84
ARCH=amd64

sudo apt install libelf-dev libncurses-dev -y

# Get sources
git clone https://git.kernel.org/pub/scm/linux/kernel/git/stable/linux.git ../linux
pushd ../linux
git checkout v${KVERSION}

# apply the configuration
cp ../config/linux-${KVERSION}.config .config

## build kernel
make -j $(nproc)
popd
```

#### Check if kernel is ready for container workloads

Use [this script](https://raw.githubusercontent.com/moby/moby/master/contrib/check-config.sh) from the [moby project](https://blog.hypriot.com/post/verify-kernel-container-compatibility/).

### Ubuntu Disk Image

The last component is a disk image with the root filesystem installed on it. Official instructions are found [here](https://www.gem5.org/documentation/general_docs/fullsystem/disks). We will use qemu to create the disk image as it allows one to configure the disk image with the serverless function and test setup before switching to actual simulation. 

The provided script `scripts/setup_image.sh` basically does the following things
1. Create disk image
2. Get Ubuntu server 20.04 image
3. Create new temporary copy to avoid redownloading 
4. Start ubuntu installation
5. Create a `root` user with password `root`
6. Add gem5 binary
7. Create gem5 service

It does this by using the cloud-init and autoinstall key. The `configs/init_disk_image.sh` is ran as a late command (after linux boots) and installs docker, golang, gem5 binary and service. This script will take a really long time, and you won't see much output. If you see three HTTP GET requests with status code 200, its working and you need to wait until the script fully terminates.

#### Gem5 binary

The [m5](https://www.gem5.org/documentation/general_docs/m5ops/) binary is a tool that allows execution of magic instructions from the running system. The tool can be used in scripts and CLI to take snapshots or exit simulation.

#### Gem5 init service

The Gem5 init service automatically starts execution of the workload 

## Resources

[Creating disk image for gem5](http://www.lowepower.com/jason/setting-up-gem5-full-system.html)
[m5 binary](https://www.gem5.org/documentation/general_docs/m5ops/)
[tutorial](https://github.com/ease-lab/vhive-asplos-tutorial/blob/main/hands-on-vHive-Gem5/setup.md)

## Setting up the simulation

### Configuring simulated hardware system

Configs are found in gem5-configs. Documentation is [here](https://www.gem5.org/documentation/learning_gem5/introduction/).
The default configured system is a dual in-order core machine. Each core has a private 16kB I and D cache and is clocked at 2.5GHz. Both cores have a common 128kB LLC and 2GB of memory.

We are only interested in the characteristics of the function itself therefore we can isolate core 1 from the rest of the system to only run the containerized function on that core using Linux's [`isolcpus` feature](https://www.oreilly.com/library/view/linux-kernel-in/0596100795/re46.html). Consequently, we can measure the workload characteristics of the containerized function without any interferences from other parts of the system. Core 0 will be used as the client that drives the function.

### Setting up containerized function

Before running the simulator, the containerized function needs to be installed onto the base image. We will use the qemu emulator as it has internet access and is faster.

#### Manual installation of containerized function

We can manually install the containerized function by using qemu and logging in as root, using the script `scripts/run_qemu.sh`.
Qemu will boot the base disk image in the `workload/` folder. As soon as the system is booted, log in as `root` with password `root`. Now we can use docker to pull the function onto the base disk image.

Change the `DISK_IMG` and `KERNEL` variables where appropriate.

```bash
#!/bin/bash
# run_qemu.sh
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT="$( cd $DIR && cd .. && pwd)"

DISK_IMG=$ROOT/workload/disk-image.img
KERNEL=$ROOT/workload/vmlinux
RAM=8G
CPUS=4

sudo qemu-system-x86_64 \
    -nographic \
    -cpu host -enable-kvm \
    -smp ${CPUS} \
    -m ${RAM} \
    -device e1000,netdev=net0 \
    -netdev type=user,id=net0,hostfwd=tcp:127.0.0.1:5555-:22  \
    -drive format=raw,file=$DISK_IMG \
    -kernel $KERNEL \
    -append 'earlyprintk=ttyS0 console=ttyS0 lpj=7999923 root=/dev/hda2'
```



```bash
docker pull vhiveease/aes-go
```

We can manually test if the function actually works by using the client we placed into the disk image to run a test. The source code of the client is found in `client/` folder.

```bash
# 1. Start your function container
# -d detaches the process and we can continue in the same console.
# -p must be set to export the ports
docker run -d --name mycontainer -p 50051:50051 vhiveease/aes-go

# run the client with the port you export in docker as well as the number of invocations you want to run.
# -addr is the address and port we where exporting with the docker command
# -n is the number of invocations the client should perform
./client -addr localhost:50051 -n 100
```

