"""
Rug Printing Factory

Scenario:
  A Rug Printing Factory has a limited number of printers and defines
  a printing processes that takes some time to print.


"""
import random

import simpy

import requests

import os

RANDOM_SEED = 42
NUM_MACHINES = 1  # Number of printers in the factory
T_INTER = 7       # Create a order every ~7 minutes
SIM_TIME = 100     # Simulation time in minutes
BUNDLE_LENGTH_FT = 10

class RugFactory(object):
    """A factory has a limited number of printers (``NUM_MACHINES``) to
    print in parallel.

    Printers will request a job from the print server, and will print. Printers have an unlimited supply of rug bundles of a set
    length (100ft). If a job requested less material then requested, then a rug fragment will be created. This rug fragment will be used
    on the next request to the print server, if the rug fragment cannot be used, it will be discarded. The rug fragmet is owned by the printer.


    """
    def __init__(self, env):
        self.env = env
        self.trash_material = 0

    def request_print_job(self, length):
        """request a print_job to be used here
            return length of material used
        """
        r = requests.get('http://localhost:8080/next', json={
            "roll_length": length,
            "include_rush": True
        })
        if r.status_code == 200:
            return r.json()["length"]
        if r.status_code == 406: # not acceptable.
            return -1
        return 0

    def new_order(self):
        # insert this order directly into the database
        pass

    def trash(self, length):
        self.trash_material+=length
        yield env.timeout(1) # it takes 1 minutes to trash a segment



def printer(env, name, rug_factory):
    """The job process (each job has a ``name``) arrives at the factory
    (``rug_factory``) and requests a cleaning machine.

    It then starts the washing process, waits for it to finish and
    leaves to never come back ...

    """
    rug_fragments = []
    print(name + ' starts up')
    while True:
        length = BUNDLE_LENGTH_FT
        # if a rug segment is available, use it
        if len(rug_fragments) > 0:
            length = rug_fragments.pop()
        pj = rug_factory.request_print_job(length)
        # if the print job has zero length, there is nothing to print. Wait a minute and try again.
        if pj == 0:
            yield env.timeout(1)
            continue
        if pj == -1:
            # throw away this segment, it's been rejected
            yield env.process(rug_factory.trash(length))
            continue
        # At this block, we have received a valid job.
        print(name + ' starts printing job of length '+ str(pj) + ' on ' + str(length))
        yield env.timeout(random.randint(pj - 2, pj + 2))
        print(name + ' completed printing job of length '+ str(pj))
        # We should save some statistics about the rug, and how long it took to print.

        # if there is a rug segment, add it to the local queue to be used on the next request
        if length-pj != 0:
            print(name + ' created a rug segment of size ' + str(length-pj))
            rug_fragments.append(length-pj)
        # We then fall out of the loop, and request a new print job from the factory

def setup(env, factory, num_machines, t_inter):
    """Create a factory, a number of initial printers and keep creating orders
    approx. every ``t_inter`` minutes."""
    # Create the Factory


    # Create 4 initial printers
    for i in range(num_machines):
        env.process(printer(env, 'Printer %d' % i, factory))

    # Create more orders while the simulation is running
    while True:
        yield env.timeout(random.randint(t_inter - 2, t_inter + 2))
        i += 1
        # Create a new Order
        factory.new_order()


# Setup and start the simulation
random.seed(RANDOM_SEED)  # This helps reproducing the results

# Create a new factory


# Create an environment and start the setup process
env = simpy.Environment()
factory = RugFactory(env)
env.process(setup(env, factory, NUM_MACHINES, T_INTER))

# Execute!
env.run(until=SIM_TIME)

print("wasted material (ft):", factory.trash_material)
