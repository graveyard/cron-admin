# cron-admin
Web interface and API for managing a cron table stored in a database. For details on cron times and conventions see [this reference](http://www.adminschoice.com/crontab-quick-reference).

## API Endpoints

#### GET: `/active-functions`

Returns a list of distinct job function names which are actively scheduled.

#### GET: `/jobs`

Query param:
* Function (Required and must be non-empty)

Returns a list of all jobs associated with the given function.

#### POST: `/jobs`

Posts a new, active job.

Body params:
* Function
* CronTime
* Workload
* TimeZone (optional, defaults to "America/Los\_Angeles")

400 errors occur when:
* Function isn't provided or is empty
* CronTime isn't provided or is an invalid format

#### DELETE: `/jobs/{job_id}`

Removes job from the database.

#### PUT: `/jobs/{job_id}`

Update the job with new values.

Body params (all required):
* Function
* IsActive
* CronTime
* Workload
* TimeZone
* Created (formatted as RFC3339)

400 errors occur when:
* Any of these fields are missing
* Function is the empty
* CronTime is invalid
* Created cannot be parsed as RFC3339

## Web interface

The web interface is a single smooth page powered by React.

### Active Jobs (Home)

Displays your active functions with an input bar for directly going to an existing or new function's details page.

### Job Details

Displays active/inactive jobs with their cron times and payloads. The interface makes it simple and easy to:

* Add a new job
* Deactivate/active an already existing job
* Delete a job
