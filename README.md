<h4 align="center">
    Working as 29 Dez 2024 !
  </h4>
  
  <h1 align="center">
    🟣 Twitch Accounts 🟣
  </h1>


  <p align="center">
   A Golang bot for creating Twitch accounts
  </p>

  Twitch Accounts provides a script for creating Twitch accounts and verify them using temporary Gmail addresses. Using the Gmail approach ensures enhanced security and reliability, making the accounts more resilient against bans and blocks. This 
  project can be useful for various purposes, such as testing or automation
  
  <h3 align="center">
    🚀 Recent Enhancements 🚀
  </h3>

  <p align="center">
    🤖 Follow bot functionality has been added! 🎉
  </p>
  
  <hr>
  
  **⭐ If you found this project helpful, illuminate it with your support by dropping a brilliant star! 🌟**
  
  ## :fire: Features
  
  ✔ Create accounts on TwitchTV
  ✔ Verify account email (Using Gmail Accounts)
  ✔ Captcha Validation
  ✔ Proxy Support
  ✔ Follow bot
  ✔ Simple and easy-to-use script

  ## ⚠️ Warning

  ❌ The follow-bot feature is currently **not working**. (since 18 Nov 2024)

  ---
  
  ## ⚙️・How to setup Twitch Accounts?
  ```sh-session
  - Basic setup
  > Clone this repository
  > Create an account on https://salamoonder.com/ then add some credits and set up Captcha Api Key on your config.go file
  > If you want to use proxy, then setup it in config.go file as well, otherwise just let the default value or "" - also, make sure that your proxy service isn't blocking access to Twitch.

  - How to run
  > Make sure you have Golang installed on your machine, then run the following command in the project root: 'go run main.go' (without the single quotes)
  > To run the follow-bot feature, use the following command: 'go run followbot/followbot.go' (without the single quotes)
  ```
  
  ## 🎉・Next Steps/Enhancements
  
  - <span style="color: gray;">~~Follow bot~~</span> <i style="color:green;">Done!</i>
  - <span style="color: gray;">~~Code cleanup~~</span>  <i style="color:green;">Done!</i>
  - <span style="color: gray;">~~Proxy configuration~~</span>  <i style="color:green;">Done!</i>
  
  ## 📄・License
  
  This project is licensed under the GPL General Public License v3.0 License - see the [LICENSE.md](./LICENSE) file for details
  ```js
    ・Educational purpose only and all your consequences caused by you actions is your responsibility
    ・Selling this Free gen is forbidden
    ・If you make a copy of this/or fork it, it must be open-source and have credits linking to this repo
  ```
  
  ## ⭐・Contributing
  Contributions are welcome! If you have any ideas, suggestions, or improvements, feel free to open an issue or create a pull request.
  
  ## ❗・Notice
  Remember, automations are against Twitch rules, do not abuse this project. I've created this tool out of genuine interest and released it for wider use. Let's keep it positive and avoid any misuse to maintain a healthy environment on Twitch.
  
  ## 💭・ChangeLog
  ```diff
    v0.0.5 ⋮ 18 nov 2024
    + Fixed 'kasada taking too long' error
    + Updated bot account creation flow to accommodate Twitch adjustments
    - Follow-bot not working

    v0.0.4 ⋮ 13 may 2024
    + Added follow-bot

    v0.0.3 ⋮ 11 may 2024
    + Added proxy support to all requests
    
    v0.0.2 ⋮ 09 may 2024
    + Code cleanup
    + Enhanced error handling
    + Improved status logging

    v0.0.1 ⋮ 07 may 2024
    + Added main script (creating accounts + email verification)
   ```
  ---
  
  <p>
    All the registered accounts information is going to be stored at results/accounts.txt
  </p>
  
  
  ## Author
  Authored by: gui-fkb [Github](https://github.com/gui-fkb)
  
  ## Credits
  This project design was based on 'twitch-account-creator' NodeJS repo by masterking32 [Github](https://github.com/masterking32). Since the project wasn't being maintained and the email verification feature proposed in the original repository had stopped working, I decided to add some of my own twist to it.
  
