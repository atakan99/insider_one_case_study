# Terms

Use each term in this one meaning. Later slices add rows.

| term | meaning | do not use for |
|---|---|---|
| brief | The free text that the marketer writes to describe a campaign. | The segment rule or the generated content. |
| segment rule | The JSON rule that selects users. The backend validates it and compiles it to SQL. | The SQL query or the list of users. |
| audience | The set of users that a segment rule selects. | The count only. Say "audience count" for the number. |
| channel | One way to send a message: SMS or email. | The provider or the queue. |
| consent | A user's permission to get messages on one channel. | General account status. |
| template | Message text with `{{vars}}`, before rendering. | A rendered message or a frontend template. |
| rendered message | A template with the variables of one user filled in. | The template. |
| campaign | One brief with its segment rule, templates, approval and sends. | One message or one send job. |
| approval | The marketer's decision that the content can be sent. | A code review or a user approval of a plan. |
| send job | The queued work to send one rendered message to one user on one channel. | One attempt or the full campaign. |
| attempt | One call to the provider for one send job. | The send job. A send job can have many attempts. |
| delivery status | The state of a send job: for example queued, sent, failed. | The HTTP status of one attempt. |
| idempotency key | The key that makes sure that the same message goes to the same user one time only. | A database primary key or a request id. |
