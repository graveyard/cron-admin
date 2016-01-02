var ActiveJob = React.createClass({
  
  onSelected: function() {
    if (this.props.onJobSelected) {
      this.props.onJobSelected(this.props.job)
    }
  },
  
  render: function() {
      return (
        <div onClick={this.onSelected}>
        <tr>
          <td> {this.props.job} </td>
        </tr>
        </div>
      )
  }
});

var ActiveJobsList = React.createClass({
  render: function() {
    var JobItems = (this.props.jobs.sort() || []).map(function(job) {
        return (
        <ActiveJob 
          key={job}
          job={job}
          onJobSelected={this.props.onJobClicked}
        />
        );
      }, this);
    return (
      <tbody>
      {JobItems}
      </tbody>
    )
  }
});

var ActiveJobs = React.createClass({
  getInitialState: function() {
    this.getActiveJobs()
    return {};
  },

  getActiveJobs: function() {
    $.ajax({
      type: "GET",
      url: "/active-jobs",
      success: function(data) {
        this.setState({jobs: data})
      }.bind(this),
      error : function(xhr, status, err) {
        api_err = xhr.responsJSON.error
        console.log("Got api error :" + api_err)
      }.bind(this)
    });
  },

  onJobClicked: function(job) {
    this.props.navigate("jobdetails", job)
  },

  addJob: function(e) {
    e.preventDefault()
    job = this.refs.jobName.getInputDOMNode().value
    this.props.navigate("jobdetails", job)
  },

  render: function() {
    if (!this.state.jobs) {
      return (
        <div> Loading jobs </div>
      )
    }

    return (
      <div>
        <div className="search-body">
          <form onSubmit={this.addJob} method="POST">
            <Input ref="jobName" name="jobName" type="text" placeholder="Search for or add a new job"/>
          </form>
        </div>
        <div className="search-body">
          <Table striped bordered>
            <ActiveJobsList 
              jobs={this.state.jobs}
              onJobClicked={this.onJobClicked}
            />
          </Table>
        </div>
      </div>
    )
  }
});
