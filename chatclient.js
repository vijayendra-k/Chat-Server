var net = require('net');
var readlineSync =require('readline-sync');
var user ={}; 
if(process.argv.length != 4){
console.log(process.argv.length);
	console.log("Usage: node %s <host> <port>", process.argv[1]);
	process.exit(1);	
}

var host=process.argv[2];
var port=process.argv[3];

if(host.length >253 || port.length >5 ){
	console.log("Invalid host or port. Try again!\nUsage: node %s <port>", process.argv[1]);
	process.exit(1);	
}
var validateRegex = /[0-9a-zA-Z]{5,}/;
var client = new net.Socket();

console.log("Simple chatclient.js developed by Vijayendra Kosigi, SECAD");
console.log("Connecting to localhost:%s", port);
console.log("Connected to: %s:%s", host, port);
console.log("You need to login before sending/receiving messages");
client.connect(port,host,connected);

function connected(){
	loginsync();
}
var authenticated = 0;
client.on("data",(data) => {

	var recvMessage = (data.toString());
	console.log("\nReceived data: " + recvMessage);
	if(!authenticated){
		if(recvMessage.includes("Invalid username or password")){
			console.log("AUTHENTICATION FAILED!! Please try again");
			loginsync();
		} else {
			console.log("AUTHENTICATION SUCCESS!!\nUsername "+ user.username);
			authenticated++;
			console.log("Welcome to the chat system.");
			console.log("Type .userlist to generate the list of active users in the system");
			console.log("Type .exit to logout and close the connection");
			console.log("Type personalmessage for private chatting");
			console.log("If you type anything other than the above mentioned instructions, then you will be directed to a public chat.")
			chatmenu()

		}
	}
})

client.on("error",(err) => {
	console.log("Error");
	process.exit(2);
})

client.on("close",(data) => {
	console.log("Connection has been disconnected");
	process.exit(3);
})

function loginsync(){

	user.username = readlineSync.question('Username: ',{


	});
	if(user.username.length<5) {
		console.log("Username must have at least 5 characters. Please try again!\n");
		loginsync();
		return;
	}
	user.password = readlineSync.question('Password: ',{
	hideEchoBack: true
	});
	if(user.password.length<5) {
		console.log("Password must have at least 5 characters. Please try again!\n");
		loginsync();
		return;
	}
	client.write(JSON.stringify(user));
}


function chatmenu(){
	const keyboard = require('readline').createInterface({
		input: process.stdin,
		output:process.stdout
	})

	keyboard.on('line', (input) => {
	if (input ===".exit"){
		setTimeout(() => {
			client.destroy();
			console.log("disconnected");
			process.exit();},1);
	}

	else if(input ===".userlist") {
		var genuserlist={}
		genuserlist.Type = "userlist";
		genuserlist.Message = "getuserlist";
		client.write(JSON.stringify(genuserlist));
	}


	else if(input ==="personalmessage") {
		var privatechat={}
		console.log("Private Chatting enabled. Please enter the username and the message to whom you want to send the message")
		privatechat.Type = "private";
		keyboard.question('To ', (toanswer) => {
			privatechat.To = toanswer;
			keyboard.question('Message ', (messageresponse) => {
				privatechat.Message = messageresponse;
				client.write(JSON.stringify(privatechat));
	 		})
		})
	}

	else {
		var publicchat={}
		publicchat.Type = "public"
		publicchat.Message  = input;
		client.write(JSON.stringify(publicchat));
		}
	})
}