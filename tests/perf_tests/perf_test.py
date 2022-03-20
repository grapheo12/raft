import os
import subprocess
from time import sleep
from datetime import datetime, timedelta
import random
from math import log
import multiprocessing as mp
import uuid

from nbformat import write

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



def client_worker(sema):
    global requests_completed

    read_log = open(f"r{os.getpid()}.log", "w")
    write_log = open(f"w{os.getpid()}.log", "w")


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
    client(400, 4000, 4)
    for p in nodes:
        p.terminate()