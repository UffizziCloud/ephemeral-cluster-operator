import matplotlib.pyplot as plt
import json
import os

data_file_prefix = 'n-simul-perf-data-'
n_simul_clusters = os.environ.get('N_SIMUL_CLUSTERS')

# Load the data from 'perf-data-PR.json'
with open(data_file_prefix+'PR.json', 'r') as file:
    data_pr = json.load(file)

# Load the data from 'perf-data-main.json'
with open(data_file_prefix+'main.json', 'r') as file:
    data_main = json.load(file)

# Convert 'time' values to integers and extract 'workers' and 'time' into separate lists for PR data
workers_pr = [item['workers'] for item in data_pr]
time_pr = [int(item['time']) for item in data_pr]

# Convert 'time' values to integers and extract 'workers' and 'time' into separate lists for main data
workers_main = [item['workers'] for item in data_main]
time_main = [int(item['time']) for item in data_main]

# Sort the data by workers to ensure correct plotting order
sorted_pr = sorted(zip(workers_pr, time_pr), key=lambda x: x[0])
sorted_main = sorted(zip(workers_main, time_main), key=lambda x: x[0])

# Unzip the sorted data
workers_pr_sorted, time_pr_sorted = zip(*sorted_pr)
workers_main_sorted, time_main_sorted = zip(*sorted_main)

# Create a plot
plt.figure(figsize=(10, 6))

# Plot sorted data from 'perf-data-PR.json'
plt.plot(workers_pr_sorted, time_pr_sorted, marker='o', linestyle='-', label='PR Branch')

# Plot sorted data from 'perf-data-main.json'
plt.plot(workers_main_sorted, time_main_sorted, marker='x', linestyle='--', label='Main Branch')

# Set the tick marks to show each worker number
plt.xticks(workers_pr_sorted)
plt.yticks(sorted(set(time_pr_sorted + time_main_sorted)))  # Combine and sort unique time values

# Add title, labels, grid, and legend
plt.title('multicluster readiness comparison: PR vs Main Branch')
plt.xlabel('number of workers (n_simultaneous_clusters='+n_simul_clusters+')')
plt.ylabel('time taken for readiness (seconds)')
plt.grid(True)
plt.legend()

# Display the plot
plt.savefig('simul_graph.png')
# plt.show()
