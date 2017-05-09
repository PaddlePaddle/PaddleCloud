from django import forms
from django.forms.extras.widgets import SelectDateWidget

import account.forms


class SignupForm(account.forms.SignupForm):
    def __init__(self, *args, **kwargs):
        super(SignupForm, self).__init__(*args, **kwargs)
        del self.fields["username"]
        self.fields.keyOrder = [
            'email',
            'password',
            'password_confirm',
            'school',
            'studentID',
            'major',
            'code']

    school = forms.CharField(max_length=256)
    studentID = forms.CharField(max_length=512)
    major = forms.CharField(max_length=256)

class SettingsForm(account.forms.SettingsForm):
    def __init__(self, *args, **kwargs):
        super(SettingsForm, self).__init__(*args, **kwargs)
        instance = getattr(self, 'instance', None)
        if instance and instance.pk:
            self.fields['email'].widget.attrs['readonly'] = True

    school = forms.CharField(max_length=256)
    studentID = forms.CharField(max_length=512)
    major = forms.CharField(max_length=256)
