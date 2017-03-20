# Simple File Server

### Access File Server
> website: **ip:port/...**

### Upload File
> website: **ip:port/upload**

### Set Access Permission for user 
> website: **ip:port/setaccess?acc=1&pwd=2&usr=3&ok=4**
- acc : The account for administrator
- pwd : The password for account
- usr : The user that you want to set
- ok  : Whether the user can access server, True mean can, false mean can't

### Set a new administrator with specific name
> website: **ip:port/setadmin?acc=1&pwd=2&name=3&ok=4**
- acc  : The account of super administrator
- pwd  : The super administrator account's password
- name : The name of administrator that you want to set
- ok   : Whether should be a administrator, true mean is, false mean is not