# Generated by Django 4.1.1 on 2022-12-07 11:50

from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
        ('printing3d', '0012_remove_sell3d_id_model_remove_sell3d_quantity_and_more'),
    ]

    operations = [
        migrations.RemoveField(
            model_name='sell3d',
            name='id_purchase',
        ),
        migrations.AddField(
            model_name='purchase',
            name='id_purchase',
            field=models.IntegerField(default=1, verbose_name='Номер заказа'),
            preserve_default=False,
        ),
    ]
