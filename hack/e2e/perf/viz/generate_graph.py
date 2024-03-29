import matplotlib.pyplot as plt
import json

# Load the data from 'perf-data-PR.json'
with open('perf-data-PR.json', 'r') as file:
    data_pr = json.load(file)

# Load the data from 'perf-data-main.json'
with open('perf-data-main.json', 'r') as file:
    data_main = json.load(file)

# Extract 'workers' and 'time' into separate lists for PR data
workers_pr = [item['workers'] for item in data_pr]
time_pr = [item['time'] for item in data_pr]

# Extract 'workers' and 'time' into separate lists for main data
workers_main = [item['workers'] for item in data_main]
time_main = [item['time'] for item in data_main]

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

# Add title, labels, grid, and legend
plt.title('Time Taken by UffizziClusters with Varying Workers')
plt.xlabel('Number of Workers')
plt.ylabel('Time')
plt.grid(True)
plt.legend()

# Save the plot as an image file
plt.savefig('simul_graph.png')