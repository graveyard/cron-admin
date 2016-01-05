var AddForm = React.createClass({

  getInitialState: function() {
    return {clicked: false, errored: false}
  },

  formSubmit: function(e) {
    e.preventDefault()
    crontime = this.refs.crontime.getInputDOMNode().value
    workload = this.refs.workload.getInputDOMNode().value
    $.ajax({
      url: "/jobs",
      type: "POST",
      data: {Function: this.props.function, CronTime: crontime, Workload: workload},
      dataType: "json",
      success: function(data) {
        this.props.getJobDetails(this.props.function)
      }.bind(this),
      error: function(xhr, status, err) {
        this.setState({errored: true, err_msg: xhr.responseText})
      }.bind(this)
    }); 
  },

  addJob: function() {
    this.setState({clicked: true})
  },

  cronAlert: function() {
    if (!this.state.errored) {
      return
    }

    return <Alert bsStyle="danger"> {this.state.err_msg} </Alert>
  },

  render: function() {
    if (!this.state.clicked) {
      return (<div><Button bsStyle="primary" onClick={this.addJob}> Add job </Button><p></p></div>)
    }

    crontime_placeholder = 'Cron time: (e.g.  0 13 */4 * * 1-5)'
    workload_placeholder = 'Workload: (e.g. "--task=job" or {task:job})'
    return (
      <div>
      {this.cronAlert()}
      <form onSubmit={this.formSubmit} method="POST">
      <Input ref="crontime" type="text" placeholder={crontime_placeholder} required />
      <Input ref="workload" type="text" placeholder={workload_placeholder}/>
      <ButtonInput bsStyle="primary" type="submit"> Submit job </ButtonInput>
      </form>
      </div>
    )
  }
});

var CronRow = React.createClass({
  getInitialState: function() {
    return {button_clicked: false, delete_clicked: false}
  },

  button_click: function() {
    if (!this.state.button_clicked) {
      this.setState({button_clicked: true})
      return
    }

    job = this.props.job
    $.ajax({
      url: "/jobs/" + job.ID,
      type: "PUT",
      data: {IsActive: !job.IsActive,
             Function: job.Function,
             CronTime: job.CronTime,
             Workload: job.Workload,
             Created: job.Created,
             TimeZone: job.TimeZone},
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

  delete_click: function() {
    if (!this.state.delete_clicked) {
      this.setState({delete_clicked: true})
      return
    }

    job = this.props.job
    $.ajax({
      url: "/jobs/" + job.ID,
      type: "DELETE",
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
    button_display = this.state.button_clicked ? "Are you sure?":display
    style = this.props.is_active ? "danger":"warning"
    delete_display = this.state.delete_clicked ? "Are you sure?":"Delete Job"
    return(
      <tr>
        <td id="button"><Button bsStyle={style} onClick={this.button_click}> {button_display} </Button></td>
        <td>{job.CronTime}</td>
        <td id="workload">{JSON.stringify(job.Workload)}</td>
        <td id="button"><Button bsStyle="danger" onClick={this.delete_click}> {delete_display} </Button></td>
      </tr>
    )
  }
});

var JobDetails = React.createClass({

  getInitialState: function() {
    this.getJobDetails(this.props.function)
    return {loading: true, jobs: []}
  },

  getJobDetails: function(func) {
    this.setState({loading: true})
    $.ajax({
      type: "GET",
      url: "/jobs",
      data: {"Function": func},
      success: function(data) {
        this.setState({loading: false, jobs: data})
      }.bind(this),
      error : function(xhr, status, err) {
        api_err = xhr.responseJSON.error
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
          <td id="buttoncol"> </td>
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
        <p id="job-name"> {this.props.function} </p>
        {this.cronUsage()}
        <AddForm function={this.props.function} getJobDetails={this.getJobDetails}/>
        {this.displayRows(true)}
        {this.displayRows(false)}
      </div>
    )
  }
});
