var CronAdmin = React.createClass({
  getUrlPathSplit: function() {
    if (window.location.hash.length > 1) {
      return window.location.hash.slice(1).split("/");
    }
    return null;
  },

  getInitialState: function() {
    var keys = this.getUrlPathSplit();
    if (!keys) {
      history.replaceState(null, "Cron-Admin", "#activefunctions");
      return {page: "activefunctions"};
    }
    return {page: keys[0], params: {}, urlParams: keys.slice(1) };
  },

  componentDidMount: function() {
    var self = this;
    window.onpopstate = function(event) {
      if (event.state) {
        self.setState({page: "activefunctions", params: {}, urlParams: null});
        return;
      }

      var keys = self.getUrlPathSplit();
      if (keys) {
        self.setState({page: keys[0], params: {}, urlParams: keys.slice(1)});
        return;
      }
      self.setState({page: "activefunctions", params: {}, urlParams: null});
    };
  },

  navClick: function(page) {
    this.navigate(page, {});
  },

  navigate: function(page, params) {
    this.setState({page: page, params: params, urlParams: null});
    history.pushState({page: page, params: params}, null, "#" + page);
  },

  render: function() {
    var mainPage;
    if (this.state.page === "activefunctions") {
      mainPage = <ActiveFunctions navigate={this.navigate} urlParams={this.state.urlParams}/>;
    } else if (this.state.page === "jobdetails") {
      mainPage = <JobDetails navigate={this.navigate} function={this.state.params} urlParams={this.state.urlParams}/>;
    } else if (this.state.page === "submitjob") {
      mainPage = <SubmitJob navigate={this.navigate} urlParams={this.state.urlParams}/>;
    }

    return (
      <div>
        <Navbar inverse fluid brand={<a href="#activefunctions">Cron Admin</a>}>
          <Nav activeKey={this.state.page}>
            <NavItem eventKey={"activefunctions"} onClick={this.navClick.bind(this, "activefunctions")} href='#activefunctions'>Main</NavItem>
            <NavItem eventKey={"submitjob"} onClick={this.navClick.bind(this, "submitjob")} href='#submitjob'>Submit a Job</NavItem>
          </Nav>
        </Navbar>
        {mainPage}
      </div>
    );
  }
});

React.render(<CronAdmin />, $("#cron-admin")[0]);
