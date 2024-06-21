from marshmallow import Schema, fields


class QuestionSchema(Schema):
    id = fields.Int(required=False)
    title = fields.Str(required=True)
    answer = fields.Str(required=True)


class QuestionEditeSchema(Schema):
    id = fields.Int(required=True)
    title = fields.Str(required=False)
    answer = fields.Str(required=False)


class ListQuestionSchema(Schema):
    questions = fields.Nested(QuestionSchema, many=True)


class SessionSchema(Schema):
    id = fields.Int(required=True)
    group_id = fields.Int(required=True)
    status = fields.Str(required=True)
    capitan_id = fields.Str(required=True)


class ListQSessionSchema(Schema):
    sessions = fields.Nested(SessionSchema, many=True)
