# Generated by Django 4.1.2 on 2022-12-27 14:53

from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
        ('printing3d', '0019_alter_user3d_password'),
    ]

    operations = [
        migrations.AddField(
            model_name='sell3d',
            name='compl_date',
            field=models.DateTimeField(auto_now=True, verbose_name='Дата продажи'),
        ),
        migrations.AddField(
            model_name='sell3d',
            name='del_date',
            field=models.DateTimeField(auto_now=True, verbose_name='Дата продажи'),
        ),
    ]
