
var CronAdmin = React.createClass({
  getUrlPathSplit: function() {
    if (window.location.hash.length > 1) {
      keys = window.location.hash.slice(1).split("/");
      return keys;
    }
    return null;
  },

  getInitialState: function() {
    keys = this.getUrlPathSplit();
    if (!keys) {
      history.replaceState(null, "Cron-Admin", "#activejobs");
      return {page: "activejobs"};
    }
    return {page: keys[0], params: {}, urlParams: keys.slice(1) };
  },

  componentDidMount: function() {
    var self = this;
    window.onpopstate = function(event) {
      if (event.state) {
        self.setState({page: "activejobs", params: {}, urlParams: null});
        return;
      }

      keys = self.getUrlPathSplit();
      if (keys) {
        self.setState({page: keys[0], params: {}, urlParams: keys.slice(1)})
        return;
      }
      self.setState({page: "activejobs", params: {}, urlParams: null})
    };
  },

  navClick: function(page) {
    this.navigate(page, {})
    return false
  },

  navigate: function(page, params) {
    this.setState({page: page, params: params, urlParams: null})
    history.pushState({page: page, params: params}, null, "#" + page)
  },

  render: function() {
    if (this.state.page === "activejobs") {
      mainPage = <ActiveJobs navigate={this.navigate} />
    } else if (this.state.page === "jobdetails") {
      mainPage = <JobDetails navigate={this.navigate} job={this.state.params} />
    }

    return (
      <div>
        <Navbar inverse fluid brand={<a href="#activejobs"> Cron Admin </a>}>
        </Navbar>
        {mainPage}
      </div>
    );
  }
});

React.render(<CronAdmin />, $("#cron-admin")[0]);
