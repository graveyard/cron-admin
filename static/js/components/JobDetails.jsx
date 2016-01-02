var AddForm = React.createClass({

  getInitialState: function() {
    return {clicked: false}
  },

  formSubmit: function(e) {
    e.preventDefault()
    crontime = this.refs.crontime.getInputDOMNode().value
    workload = this.refs.workload.getInputDOMNode().value
    $.ajax({
      url: "/add-job",
      type: "POST",
      data: {job: this.props.job,crontime: crontime, workload: workload},
      dataType: "json",
      success: function(data) {
        console.log("Success!")
        this.props.getJobDetails(this.props.job)
      }.bind(this),
      error: function(xhr, status, err) {
        api_error = xhr.responseJSON.error
        console.log("Error " + api_error)
      }.bind(this)
    }); 
  },

  addJob: function() {
    this.setState({clicked: true})
  },

  render: function() {
    if (!this.state.clicked) {
      return (<div><Button bsStyle="primary" onClick={this.addJob}> Add job </Button><p></p></div>)
    }

    crontime_placeholder = 'Cron time: (e.g.  0 13 */4 * * 1-5)'
    workload_placeholder = 'Workload: (e.g. "--task=job" or {"task":"job"})'
    return (
      <form onSubmit={this.formSubmit} method="POST">
      <Input ref="crontime" type="text" placeholder={crontime_placeholder} required />
      <Input ref="workload" type="text" placeholder={workload_placeholder}/>
      <ButtonInput bsStyle="primary" type="submit"> Submit job </ButtonInput>
      </form>
    )
  }
});

var CronRow = React.createClass({
  getInitialState: function() {
    return {clicked: false}
  },

  click: function() {
    if (!this.state.clicked) {
      this.setState({clicked: true})
      return
    }

    $.ajax({
      url: "/modify-job/" + this.props.job.ID,
      type: "POST",
      data: {active: !this.props.is_active},
      dataType: "json",
      success: function(data) {
        this.props.getJobDetails(this.props.job.Function)
      }.bind(this),
      error: function(xhr, status, err) {
        api_error = xhr.responseJSON.error
        console.log("Error " + api_error)
      }.bind(this)
    }); 
  },

  render: function() {
    job = this.props.job
    display = this.props.is_active ? "Deactivate":"Activate"
    button_display = this.state.clicked ? "Are you sure?":display
    style = this.props.is_active ? "danger":"warning"
    return(
      <tr>
        <td id="button"><Button bsStyle={style} onClick={this.click}> {button_display} </Button></td>
        <td>{job.CronTime}</td>
        <td id="workload">{JSON.stringify(job.Workload)}</td>
      </tr>
    )
  }
});

var JobDetails = React.createClass({

  getInitialState: function() {
    this.getJobDetails(this.props.job)
    return {loading: true, jobs: []}
  },

  getJobDetails: function(job) {
    this.setState({loading: true})
    $.ajax({
      type: "GET",
      url: "/job-details",
      data: {"job": job},
      success: function(data) {
        this.setState({loading: false, jobs: data})
      }.bind(this),
      error : function(xhr, status, err) {
        api_err = xhr.responsJSON.error
        this.setState({loading: false})
        console.log("Got api error :" + api_err)
      }.bind(this)
    });
  },

  displayRows: function(is_active) {
    if (!this.state.jobs) {
      return
    }

    jobs = this.state.jobs.sort(function asc_sort(a, b) {
          return (b.Workload < a.Workload ? 1 : -1)    
    })
    title = is_active ? "Active jobs":"Inactive jobs"
    header = (
        <tr>
          <td id="buttoncol"> </td>
          <td id="croncol"> Cron time </td>
          <td> Workload </td>
        </tr>
    )

    rows = []
    for (i in jobs) {
      job = jobs[i]
      if (job.IsActive != is_active) {
        continue
      }
      row=<CronRow key={job.ID} job={job} is_active={is_active} getJobDetails={this.getJobDetails}/>
      rows.push(row)
    }
    if (rows.length < 1) {
      return 
    }
    return (
        <div>
          <p id="active"> {title} ({rows.length}) </p>
          <Table striped bordered> 
            <thead> {header} </thead>
            <tbody> {rows} </tbody>
          </Table>
        </div>
      )
  },

  cronUsage: function() {
    return (
      <Alert bsStyle="info"> For more information on cron convention please see <a href="https://github.com/ncb000gt/node-cron"> the README to the node package clever-cron uses. </a></Alert>
    )
  },

  render: function() {
    if (this.state.loading) {
      return (<p> Loading... </p>)
    }

    return (
      <div className="job-details">
        <p id="job-name"> {this.props.job} </p>
        {this.cronUsage()}
        <AddForm job={this.props.job} getJobDetails={this.getJobDetails}/>
        {this.displayRows(true)}
        {this.displayRows(false)}
      </div>
    )
  }
});
