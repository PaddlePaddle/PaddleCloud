from django import forms
from django.forms.extras.widgets import SelectDateWidget

import account.forms


class SignupForm(account.forms.SignupForm):
    def __init__(self, *args, **kwargs):
        super(SignupForm, self).__init__(*args, **kwargs)
        del self.fields["username"]
    school = forms.CharField(max_length=256)
    studentID = forms.CharField(max_length=512)
    major = forms.CharField(max_length=256)

class SettingsForm(account.forms.SettingsForm):

    school = forms.CharField(max_length=256)
    studentID = forms.CharField(max_length=512)
    major = forms.CharField(max_length=256)
