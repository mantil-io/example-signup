## About

This example shows how to create a simple signup workflow using Mantil. It consists of two steps:
- **Registration** - the user registers using a valid email address and receives an activation code
- **Activation** - the user uses the activation code to confirm the registration. In return, they receive a JWT token which they can then use to authenticate.

## Prerequisites

This example is created with Mantil. To download [Mantil CLI](https://github.com/mantil-io/mantil#installation) on Mac or Linux use Homebrew 
```
brew tap mantil-io/mantil
brew install mantil
```
or check [direct download links](https://github.com/mantil-io/mantil#installation).

To deploy this application you will need an [AWS account](https://aws.amazon.com/premiumsupport/knowledge-center/create-and-activate-aws-account/).

## Installation

To locally create a new project from this example run:
```
mantil new app --from https://github.com/mantil-io/example-signup
cd app
```

## Configuration

In order for mailing to work properly, you will need to configure some environment variables:
- **APP_NAME** - the name of your application, this can be any string as it is only used to generate the mail body
- **SOURCE_MAIL** - the source email address that all mails are sent from. In order to send emails from this address, you must first verify it using the SES console. You can find step-by-step instructions for this [here](https://aws.amazon.com/getting-started/hands-on/send-an-email/). Note that if your account is in the SES sandbox you must also verify all recipient's email addresses or [apply](https://docs.aws.amazon.com/ses/latest/DeveloperGuide/request-production-access.html) to move out of the sandbox.

Finally, here is an example configuration:
```
project:
  stages:
    - name: development
      functions:
      - name: signup
        env:
          APP_NAME: Mantil
          SOURCE_MAIL: hello@mantil.com
```

## Deploying the application

Note: If this is the first time you are using Mantil you will need to install Mantil Node on your AWS account. For detailed instructions please follow the [one-step setup](https://github.com/mantil-io/mantil/blob/master/docs/getting_started.md#setup)
```
mantil aws install
```
Then you can proceed with application deployment.
```
mantil deploy
```
This command will create a new stage for your project with the default name `development` and deploy it to your node.

You will now have access to three methods:
- `register` - expects a valid email address. Creates a registration record in dynamodb and sends a mail to the user containing an activation code.
- `activate` - expects a valid activation code. Creates an activation record in dynamodb and returns a valid JWT token containing the activation ID and code.
- `verify` - expects a JWT token returned by `activate`. Checks if the JWT token is valid and returns the decoded token if successful.

An example signup flow might look like this:

First, we invoke the `register` method:
```
mantil invoke signup/register -d '{"email":"daniel@mantil.com","name":"Daniel"}'
```
After receiving the mail with the activation code, we use it to invoke the `activate` method:
```
mantil invoke signup/activate -d '{"activationCode":"b74012e1-0137-445a-bcb5-47842a9efa3e"}'
```
This returns a token that we can verify using the `verify` method:
```
mantil invoke signup/verify -d '{"token":"eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJhY3RpdmF0aW9uQ29kZSI6ImI3NDAxMmUxLTAxMzctNDQ1YS1iY2I1LTQ3ODQyYTllZmEzZSIsImFjdGl2YXRpb25JRCI6ImM3ZDMwOTIwLTE5NmQtNGViZi04ZTMwLTFjMWJjYjViOGUwMSIsImNyZWF0ZWRBdCI6MTY0MTkwNjM4NTQwMSwiaWF0IjoxNjQxOTA2Mzg1LCJleHAiOjE2NzM0NDIzODV9.kgtrAJ4Wm3DkjVdbH_cTm576LsD9GZG8P4zmVbDrCVJSUueIsx_RIJ0oKPSag569D4fzbpz-JF_dSSnlPvZ7BA"}'
```
which returns the decoded token:
```
200 OK
{
   "activationCode": "b74012e1-0137-445a-bcb5-47842a9efa3e",
   "activationID": "c7d30920-196d-4ebf-8e30-1c1bcb5b8e01",
   "createdAt": 1641906385401
}
```

## Cleanup

To remove the created stage from your AWS account destroy it with:
```
mantil stage destroy development
```

## Final thoughts

This example uses Mantil's persistent key/value storage that you can learn more about in the [todo example](https://github.com/mantil-io/example-todo).

If you have any questions or comments on this template or would just like to share your view on Mantil contact us at [support@mantil.com](mailto:support@mantil.com).
