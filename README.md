# shift-scheduler
This is a Go-HTMX project for the Full Stack Training at BYU OIT. It will allow a user to schedule out their shifts.


## How to run the app

After cloning the app, open a terminal and navigate to the directory where the project has been cloned then run `go mod tidy` so any necessary dependencies can be installed.

Once this has finished, run `go run "./cmd/webserver/."` in the terminal to start the app.

The app will start up on localhost:8081. Open a browser and navigate to that link. 

Once you are done testing the app, you can return to the terminal and use Ctrl + C to stop the app running. 

## User Flow

There are two regular users already populated in users.json, student1 and student2. 

The set password for student1 is userpass. The password for student2 is happen

Login as one of the users. 

You will be automatically redirected to the schedule page where you can submit a schedule. 


## Admin Flow

There are two admin users already populated in users.json, admin1 and admin2. 

The set password for student1 is userpass. The password for student2 is happen.

Login as one of the admins. 

You will be automatically redirected to the schedule page where you can see all pending schedules or, if no schedules have been submitted, the page has a note that there are no schedules for review. Test out the functionality of approving or rejecting a schedule.