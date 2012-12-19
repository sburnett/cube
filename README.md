cube
====

Send Go expvars to Cube for visualization.

Pretty documentation at http://godoc.org/github.com/sburnett/cube

Because expvars are usually monotonic counters, it's often useful to plot their
change over the past minute. If you're using Cubism to plot your data, you can
do this by subtracting the previous minute's counter:

	function metricPerSecond(metric) {
	    var current = cube.metric(metric),
	        shift = current.shift(-60 * 1000),
	        change = current.subtract(shift).divide(60);
	    return change;
	}

	// ... Set up your graphs ...
	var allocs = context.horizon().metric(metricPerSecond("max(myevents(memstats.Mallocs))"))
								  .title("memstats.MallocsPerSecond");
