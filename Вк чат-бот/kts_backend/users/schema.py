from marshmallow import Schema, fields


class UserSchema(Schema):
    id = fields.Int(required=True)
    first_name = fields.Str(required=False)
    last_name = fields.Str(required=False)
