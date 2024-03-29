import matplotlib.pyplot as plt

# Cold start performance data in seconds
cold_start_main = 120  # Replace with actual cold start time for the main branch
cold_start_pr = 150  # Replace with actual cold start time for the PR branch

# Branch names
branches = ['Main Branch', 'PR Branch']

# Performance times
performance_times = [cold_start_main, cold_start_pr]

# Create a bar chart
plt.figure(figsize=(8, 6))
plt.bar(branches, performance_times, color=['blue', 'orange'])

# Add a grid
plt.grid(True, linestyle='--', which='both', axis='y', alpha=0.7)

# Set the y-axis ticks to align with the actual values
plt.yticks(range(0, max(performance_times) + 10, 10))

# Add title and labels
plt.title('Cold Start Performance Comparison : PR vs Main Branch')
plt.xlabel('Branch')
plt.ylabel('Time Taken (seconds)')

# Show the plot

plt.savefig('cold_start_graph.png')
# plt.show()