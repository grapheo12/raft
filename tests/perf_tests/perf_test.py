import os
import subprocess
from time import sleep
from datetime import datetime, timedelta
import random
from math import log
import multiprocessing as mp
import uuid

from collections import defaultdict
import re

from sys import argv


mp.set_start_method("fork")

nodes = []
def spawn_nodes(n):
    for i in range(n):
        f = open(f"log{str(i)}.log", "w")
        p = subprocess.Popen(["./bin/raft",
        "-c", "tests/perf_tests/config.ini",
        "-l", "25",
        "--id", str(i),
        "--network-port", str(3001 + i),
        "--client-port",  str(4001 + i)],
        stdout=f
        )

        nodes.append(p)


requests_completed = mp.Value('i')


def read_request(log):
    global nodes
    i = random.randint(0, len(nodes)-1)

    start = datetime.now()
    p = subprocess.Popen(["./bin/client", "read",
    "-a", f"127.0.0.1:{str(4001 + i)}",
    ],
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE
    )

    stdout, _ = p.communicate()
    end = datetime.now()

    elapsed = end - start

    log.write(stdout.decode() + str(elapsed.microseconds) + "\n\n")


def write_request(log):
    global nodes
    i = random.randint(0, len(nodes)-1)

    msg = uuid.uuid4()

    start = datetime.now()
    p = subprocess.Popen(["./bin/client", "write",
    "-a", f"127.0.0.1:{str(4001 + i)}",
    "-m", str(msg)
    ],
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE
    )

    stdout, _ = p.communicate()
    end = datetime.now()

    elapsed = end - start

    log.write(stdout.decode() + str(elapsed.microseconds) + "\n\n")


read_log_files = mp.Queue(maxsize=10)
write_log_files = mp.Queue(maxsize=10)

def client_worker(sema):
    global requests_completed

    read_log = open(f"r{os.getpid()}.log", "w")
    write_log = open(f"w{os.getpid()}.log", "w")
    read_log_files.put(f"r{os.getpid()}.log")
    write_log_files.put(f"w{os.getpid()}.log")


    while True:
        sema.acquire()

        with requests_completed.get_lock():
            requests_completed.value += 1

        rw = random.randint(0, 1) * random.randint(0, 1)
        if rw == 0: # Read
            read_request(read_log)
        else: # Write
            write_request(write_log)

        


def client(rate, n, workers):
    # rate in requests per second
    arrivals = []
    curr = datetime.now()

    for _ in range(n):
        arrivals.append(curr)

        # Inverse Transform Sampling
        u = random.uniform(0.0, 0.9999)
        p = -log(1 - u) * (1 / rate)
        waitTime = timedelta(seconds=p)
        curr += waitTime

    procs = []
    sema = mp.Semaphore(value=0)
    requests_completed.value = 0
    for _ in range(workers):
        p = mp.Process(target=client_worker, args=(sema,))
        procs.append(p)
        p.start()

    
    i = 0
    while i < len(arrivals):
        if datetime.now() >= arrivals[i]:
            sema.release()
            i += 1
    
    while True:
        with requests_completed.get_lock():
            if requests_completed.value >= n:
                break

    for p in procs:
        p.terminate()
        p.join()



if __name__ == "__main__":
    spawn_nodes(5)
    sleep(2)
    client(int(argv[1]), 4400, 5)
    for p in nodes:
        p.terminate()

    read_rgx = re.compile(r"^([0-9]+)$", flags=re.MULTILINE)
    read_rgx2 = re.compile(r"(\[.*\])", flags=re.MULTILINE)

    read_times = []
    read_arrs = []
    while read_log_files.qsize() != 0:
        rf = read_log_files.get()
        with open(rf, "r") as f:
            lines = f.read()
            read_times.extend(read_rgx.findall(lines))
            read_arrs.extend(read_rgx2.findall(lines))


    read_times = [int(r) for r in read_times]

    read_checker = defaultdict(lambda: set())

    for r in read_arrs:
        r = r.strip()
        sz = len(r[1:-1].split())
        read_checker[sz].add(r)

    print("Mean read latency:", sum(read_times)/len(read_times), "us", "\tRead Throughput:", len(read_arrs) / sum(read_times) * 1e+6, "req/s")
    
    for v in read_checker.values():
        assert len(v) == 1

    print("Reads verified for correctness")


    write_times = []
    good_writes = []
    write_rgx = re.compile(r"^(Write by .*)$", flags=re.MULTILINE)
    while write_log_files.qsize() != 0:
        rf = write_log_files.get()
        with open(rf, "r") as f:
            lines = f.read()
            write_times.extend(read_rgx.findall(lines))
            good_writes.extend(write_rgx.findall(lines))

    write_times = [int(w) for w in write_times]
    
    print("Mean write latency:", sum(write_times)/len(write_times), "us", "\tWrite Throughput:", len(good_writes) / sum(write_times) * 1e+6, "req/s")