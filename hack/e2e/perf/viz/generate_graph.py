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

# Create a plot
plt.figure(figsize=(10, 6))

# Plot data from 'perf-data-PR.json'
plt.plot(workers_pr, time_pr, marker='o', linestyle='-', label='PR Branch')

# Plot data from 'perf-data-main.json'
plt.plot(workers_main, time_main, marker='x', linestyle='--', label='Main Branch')

# Add title, labels, grid, and legend
plt.title('Time Taken by UffizziClusters with Varying Workers')
plt.xlabel('Number of Workers')
plt.ylabel('Time')
plt.grid(True)
plt.legend()

# Save the plot as an image file
plt.savefig('simul_graph.png')